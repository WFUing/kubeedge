/*
Copyright 2019 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package edgehub

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/kubeedge/beehive/pkg/common"
	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/mocks/edgehub"
	module "github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/edgehub/config"
)

func init() {
	add := &common.ModuleInfo{
		ModuleName: module.EdgeHubModuleName,
		ModuleType: common.MsgCtxTypeChannel,
	}

	beehiveContext.InitContext([]string{common.MsgCtxTypeChannel})

	beehiveContext.AddModule(add)
	beehiveContext.AddModuleGroup(module.EdgeHubModuleName, module.EdgeHubModuleName)
}

// TestIsSyncResponse() tests whether there exists a channel with the given message_id in the syncKeeper
func TestIsSyncResponse(t *testing.T) {
	tests := []struct {
		name  string
		hub   *EdgeHub
		msgID string
		want  bool
	}{
		{
			name:  "Sync message response case",
			hub:   &EdgeHub{},
			msgID: "test",
			want:  true,
		},
		{
			name:  "Non sync message response  case",
			hub:   &EdgeHub{},
			msgID: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSyncResponse(tt.msgID); got != tt.want {
				t.Errorf("TestController_isSyncResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDispatch() tests whether the messages are properly dispatched to their respective modules
func TestDispatch(t *testing.T) {
	tests := []struct {
		name          string
		hub           *EdgeHub
		message       *model.Message
		expectedError error
		isResponse    bool
	}{
		{
			name:          "dispatch with valid input",
			hub:           &EdgeHub{},
			message:       model.NewMessage("").BuildRouter(module.EdgeHubModuleName, module.TwinGroup, "", ""),
			expectedError: nil,
			isResponse:    false,
		},
		{
			name:          "Error Case in dispatch",
			hub:           &EdgeHub{},
			message:       model.NewMessage("test").BuildRouter(module.EdgeHubModuleName, module.EdgedGroup, "", ""),
			expectedError: fmt.Errorf("failed to handle message, no handler found for the message, message group: edged"),
			isResponse:    true,
		},
		{
			name:          "Response Case in dispatch",
			hub:           &EdgeHub{},
			message:       model.NewMessage("test").BuildRouter(module.EdgeHubModuleName, module.TwinGroup, "", ""),
			expectedError: nil,
			isResponse:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.hub.dispatch(*tt.message)
			if !reflect.DeepEqual(err, tt.expectedError) {
				t.Errorf("TestController_dispatch() error = %v, wantErr %v", err, tt.expectedError)
			}
		})
	}
}

// TestRouteToEdge() is used to test whether the message received from websocket is dispatched to the required modules
func TestRouteToEdge(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAdapter := edgehub.NewMockAdapter(mockCtrl)
	hub := newEdgeHub(true)
	hub.chClient = mockAdapter

	tests := []struct {
		name         string
		hub          *EdgeHub
		receiveTimes int
	}{
		{
			name:         "Route to edge with proper input",
			hub:          hub,
			receiveTimes: 0,
		},
		{
			name:         "Receive Error in route to edge",
			hub:          hub,
			receiveTimes: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter.EXPECT().Receive().Return(*model.NewMessage("test").BuildRouter(module.EdgeHubModuleName, module.EdgedGroup, "", ""), nil).Times(tt.receiveTimes)
			mockAdapter.EXPECT().Receive().Return(*model.NewMessage("test").BuildRouter(module.EdgeHubModuleName, module.TwinGroup, "", ""), nil).Times(tt.receiveTimes)
			mockAdapter.EXPECT().Receive().Return(*model.NewMessage(""), errors.New("Connection Refused")).Times(1)
			go tt.hub.routeToEdge()
			stop := <-tt.hub.reconnectChan
			if stop != struct{}{} {
				t.Errorf("TestRouteToEdge error got: %v want: %v", stop, struct{}{})
			}
		})
	}
}

// TestSendToCloud() tests whether the send to cloud functionality works properly
func TestSendToCloud(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAdapter := edgehub.NewMockAdapter(mockCtrl)

	msg := model.NewMessage("").BuildHeader("test_id", "", 1)
	msg.Header.Sync = true
	tests := []struct {
		name            string
		hub             *EdgeHub
		message         model.Message
		expectedError   error
		mockError       error
		HeartbeatPeriod int32
	}{
		{
			name: "send to cloud with proper input",
			hub: &EdgeHub{
				chClient: mockAdapter,
			},
			HeartbeatPeriod: 6,
			message:         *msg,
			expectedError:   nil,
			mockError:       nil,
		},
		{
			name: "Wait Error in send to cloud",
			hub: &EdgeHub{
				chClient: mockAdapter,
			},
			HeartbeatPeriod: 3,
			message:         *msg,
			expectedError:   nil,
			mockError:       nil,
		},
		{
			name: "Send Failure in send to cloud",
			hub: &EdgeHub{
				chClient: mockAdapter,
			},
			HeartbeatPeriod: 3,
			message:         model.Message{},
			expectedError:   fmt.Errorf("failed to send message, error: Connection Refused"),
			mockError:       errors.New("Connection Refused"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter.EXPECT().Send(gomock.Any()).Return(tt.mockError).Times(1)
			config.Config.Heartbeat = tt.HeartbeatPeriod
			err := tt.hub.sendToCloud(tt.message)
			if !reflect.DeepEqual(err, tt.expectedError) {
				t.Errorf("SendToCloud() error = %v, wantErr %v", err, tt.expectedError)
			}
		})
	}
}

// TestRouteToCloud() tests the reception of the message from the beehive framework and forwarding of that message to cloud
func TestRouteToCloud(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAdapter := edgehub.NewMockAdapter(mockCtrl)
	config.Config.MessageQPS = 3
	config.Config.MessageBurst = 6
	hub := newEdgeHub(true)
	hub.chClient = mockAdapter

	tests := []struct {
		name string
		hub  *EdgeHub
	}{
		{
			name: "Route to cloud with valid input",
			hub:  hub,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter.EXPECT().Send(gomock.Any()).Return(errors.New("Connection Refused")).AnyTimes()

			core.Register(&EdgeHub{enable: true})

			go tt.hub.routeToCloud()
			time.Sleep(2 * time.Second)

			msg := model.NewMessage("").BuildHeader("test_id", "", 1)
			beehiveContext.Send(module.EdgeHubModuleName, *msg)
			stopChan := <-tt.hub.reconnectChan
			if stopChan != struct{}{} {
				t.Errorf("Error in route to cloud")
			}
		})
	}
}

// TestKeepalive() tests whether ping message sent to the cloud at regular intervals happens properly
func TestKeepalive(t *testing.T) {
	CertFile := "/tmp/kubeedge/certs/edge.crt"
	KeyFile := "/tmp/kubeedge/certs/edge.key"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAdapter := edgehub.NewMockAdapter(mockCtrl)
	tests := []struct {
		name string
		hub  *EdgeHub
	}{
		{
			name: "Heartbeat failure Case",
			hub: &EdgeHub{
				chClient:      mockAdapter,
				reconnectChan: make(chan struct{}),
			},
		},
	}
	edgeHubConfig := config.Config
	edgeHubConfig.TLSCertFile = CertFile
	edgeHubConfig.TLSPrivateKeyFile = KeyFile

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter.EXPECT().Send(gomock.Any()).Return(nil).Times(1)
			mockAdapter.EXPECT().Send(gomock.Any()).Return(errors.New("Connection Refused")).Times(1)
			go tt.hub.keepalive()
			got := <-tt.hub.reconnectChan
			if got != struct{}{} {
				t.Errorf("TestKeepalive() StopChan = %v, want %v", got, struct{}{})
			}
		})
	}
}

// 模拟 exec.Command 的函数
var mockExecCommand func(command string, args ...string) *exec.Cmd

// 用于替换 exec.Command 的假命令执行器
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// 辅助测试进程，用于模拟实际的外部命令执行
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// 模拟 benchmark.sh 输出
	output := `{
		"cpu_benchmark": 2198.53,
		"memory_benchmark": 8953271.26,
		"disk_io_benchmark": 13755120.0
	}`
	// 输出假定的 JSON 结果
	os.Stdout.Write([]byte(output))
	os.Exit(0)
}

// probe 方法的单元测试
func TestProbe(t *testing.T) {
	// 保存原始 exec.Command 并在测试结束时恢复
	originalExecCommand := exec.Command
	defer func() { mockExecCommand = originalExecCommand }()

	// 替换为假命令执行器
	mockExecCommand = fakeExecCommand

	// 创建 EdgeHub 实例，并模拟 chClient
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAdapter := edgehub.NewMockAdapter(mockCtrl)

	eh := &EdgeHub{
		chClient:      mockAdapter, // 模拟 chClient
		reconnectChan: make(chan struct{}),
	}

	// 模拟 send 方法的行为，不做任何操作即可
	mockAdapter.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

	// 使用超时上下文来确保测试不进入死循环
	done := make(chan bool)
	go func() {
		// 执行探针方法
		eh.probe()
		done <- true
	}()

	select {
	case <-done:
		// 测试成功完成
	case <-time.After(100 * time.Second):
		// 测试超时
		t.Fatal("TestProbe timed out")
	}
}
