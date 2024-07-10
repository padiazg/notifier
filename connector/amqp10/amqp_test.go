package amqp10

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/padiazg/notifier/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInternalWrapper struct {
	mock.Mock
	Config *Config
	wait   time.Duration
}

func (m *MockInternalWrapper) Dial(ctx context.Context, addr string, opts *amqp.ConnOptions) error {
	args := m.Called(ctx, addr, opts)
	return args.Error(0)
}

func (m *MockInternalWrapper) NewSession(ctx context.Context, opts *amqp.SessionOptions) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *MockInternalWrapper) NewSender(ctx context.Context, target string, opts *amqp.SenderOptions) error {
	args := m.Called(ctx, target, opts)
	return args.Error(0)
}

func (m *MockInternalWrapper) Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error {
	args := m.Called(ctx, msg, opts)
	if m.wait > 0 {
		fmt.Printf("waiting %d\n", m.wait)
		time.Sleep(m.wait)
	}

	return args.Error(0)
}

func (m *MockInternalWrapper) CloseConn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockInternalWrapper) CloseSession(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func checkName(name string) model.TestCheckNotifierFn {
	return func(t *testing.T, np model.Notifier) {
		t.Helper()
		an, _ := np.(*AMQPNotifier)
		if name != "" {
			assert.Equalf(t, name, an.Name(), "checkName = %s, expected %s", an.Name(), name)
			return
		}

		var (
			test = an.Type() + `[abcdef0-9]{8}`
			re   = regexp.MustCompile(test)
		)

		assert.Regexpf(t, re, an.Name(), "checkName = %v doesn't appply for %v", an.Name(), test)
	}
}

func TestAMQPNotifier_New(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test-logger", log.LstdFlags)

		checkConfigSet = func() model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*AMQPNotifier)
				assert.NotNilf(t, n.Config, "configSet Config expexted to be not nil")
			}
		}

		checkName = func(name string) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				n, _ := np.(*AMQPNotifier)
				if name != "" {
					assert.Equalf(t, name, n.Name(), "checkName = %s, expected %s", n.Name(), name)
					return
				}

				var (
					test = n.Type() + `[abcdef0-9]{8}`
					re   = regexp.MustCompile(test)
				)

				assert.Regexpf(t, re, n.Name(), "checkName = %v doesn't appply for %v", n.Name(), test)
			}
		}

		checkLogger = func(mark string) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				n, _ := np.(*AMQPNotifier)

				if assert.NotEmptyf(t, n.Logger, "checkLogger Logger is empty, expected to be set") {
					n.Logger.Printf(mark)
					if got := buf.String(); !strings.Contains(got, mark) {
						t.Errorf("checkLogger log = %s, expected %s", got, mark)
					}
				}
			}
		}

		checkCtx = func() model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				n, _ := np.(*AMQPNotifier)
				assert.NotNilf(t, n.ctx, "ctx is nil, expected not to")
			}
		}

		checkWrapper = func() model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				n, _ := np.(*AMQPNotifier)
				assert.NotNilf(t, n.wrapper, "wrapper is nil, expected not to")
			}
		}

		checkChannel = func(wantPanic bool) model.TestCheckNotifierFn {
			return func(t *testing.T, np model.Notifier) {
				t.Helper()
				an, _ := np.(*AMQPNotifier)

				defer func() {
					if r := recover(); r != nil && !wantPanic {
						t.Errorf("chechChannel has panicked, expected not to")
					}
				}()

				go func() {
					close(an.Channel)
				}()

				time.Sleep(20 * time.Millisecond)
				select {
				case <-an.Channel:
				default:
					t.Errorf("chechChannel is not closed")
				}
			}
		}

		tests = []struct {
			name   string
			config *Config
			checks []model.TestCheckNotifierFn
		}{
			{
				name:   "success-empty-config",
				config: nil,
				checks: model.CheckNotifier(
					checkConfigSet(),
					checkName(""),
					checkLogger(""),
					checkCtx(),
					checkWrapper(),
					checkChannel(false),
				),
			},
			{
				name: "success-full-config",
				config: &Config{
					Name:      "amqp-10",
					QueueName: "test-queue",
					Address:   "amqp://localhost",
					Logger:    logger,
				},
				checks: model.CheckNotifier(
					checkConfigSet(),
					checkName("amqp-10"),
					checkLogger("test-logger-a"),
				),
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)
			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}

func TestAMQPNotifier_Type(t *testing.T) {
	var (
		n    = New(&Config{})
		got  = n.Type()
		want = "amqp10"
	)

	assert.Equalf(t, got, want, "Type() = %v, want %s", got, want)
}

func TestAMQPNotifier_Name(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []model.TestCheckNotifierFn
	}{
		{
			name:   "success-no-name-set",
			config: nil,
			checks: model.CheckNotifier(
				checkName(""),
			),
		},
		{
			name:   "success-name-set",
			config: &Config{Name: "amqp-01"},
			checks: model.CheckNotifier(
				checkName("amqp-01"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := (&AMQPNotifier{}).New(tt.config)
			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}

func TestAMQPNotifier_Connect(t *testing.T) {
	var (
		buf bytes.Buffer

		tests = []struct {
			name       string
			wantErrMsg string
			before     func(n *AMQPNotifier)
		}{
			{
				name: "success",
				before: func(n *AMQPNotifier) {
					w := n.wrapper.(*MockInternalWrapper)
					w.On("Dial", n.ctx, "amqp://example.com", (*amqp.ConnOptions)(nil)).
						Return(nil)
					w.On("NewSession", n.ctx, (*amqp.SessionOptions)(nil)).
						Return(nil)
					w.On("NewSender", n.ctx, "test-queue", (*amqp.SenderOptions)(nil)).
						Return(nil)
				},
				wantErrMsg: "",
			},
			{
				name: "fail-Dial",
				before: func(n *AMQPNotifier) {
					w := n.wrapper.(*MockInternalWrapper)
					w.On("Dial", n.ctx, "amqp://example.com", (*amqp.ConnOptions)(nil)).
						Return(fmt.Errorf("test-Dial-error"))
					w.On("NewSession", n.ctx, (*amqp.SessionOptions)(nil)).
						Return(nil)
					w.On("NewSender", n.ctx, "test-queue", (*amqp.SenderOptions)(nil)).
						Return(nil)
				},
				wantErrMsg: "dialing AMQP server:",
			},
			{
				name: "fail-NewSession",
				before: func(n *AMQPNotifier) {
					w := n.wrapper.(*MockInternalWrapper)
					w.On("Dial", n.ctx, "amqp://example.com", (*amqp.ConnOptions)(nil)).
						Return(nil)
					w.On("NewSession", n.ctx, (*amqp.SessionOptions)(nil)).
						Return(fmt.Errorf("test-NewSession-error"))
					w.On("NewSender", n.ctx, "test-queue", (*amqp.SenderOptions)(nil)).
						Return(nil)
				},
				wantErrMsg: "creating AMQP session:",
			},
			{
				name: "fail-NewSender",
				before: func(n *AMQPNotifier) {
					w := n.wrapper.(*MockInternalWrapper)
					w.On("Dial", n.ctx, "amqp://example.com", (*amqp.ConnOptions)(nil)).
						Return(nil)
					w.On("NewSession", n.ctx, (*amqp.SessionOptions)(nil)).
						Return(nil)
					w.On("NewSender", n.ctx, "test-queue", (*amqp.SenderOptions)(nil)).
						Return(fmt.Errorf("test-NewSender-error"))
				},
				wantErrMsg: "creating sender link:",
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				n = New(&Config{
					Address:   "amqp://example.com",
					QueueName: "test-queue",
					ctx:       context.TODO(),
					wrapper:   &MockInternalWrapper{},
					Logger:    log.New(&buf, "test-logger", log.LstdFlags),
				})

				wantErr = tt.wantErrMsg != ""
			)

			if tt.before != nil {
				tt.before(n)
			}

			err := n.Connect()
			if wantErr {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAMQPNotifier_Close(t *testing.T) {
	var (
		buf   bytes.Buffer
		tests = []struct {
			name       string
			wantErrMsg string
			before     func(n *AMQPNotifier)
		}{
			{
				name: "success",
				before: func(n *AMQPNotifier) {
					w := n.wrapper.(*MockInternalWrapper)
					w.On("CloseConn").Return(nil)
				},
				wantErrMsg: "",
			},
			{
				name: "fail",
				before: func(n *AMQPNotifier) {
					w := n.wrapper.(*MockInternalWrapper)
					w.On("CloseConn").Return(fmt.Errorf("wrapper-CloseConn-error"))
				},
				wantErrMsg: "wrapper-CloseConn-error",
			},
			{
				name: "fail-wrapper-empty",
				before: func(n *AMQPNotifier) {
					n.Config.wrapper = nil
				},
				wantErrMsg: "can't call Close, Wrapper not set",
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				n = New(&Config{
					Address:   "amqp://example.com",
					QueueName: "test-queue",
					ctx:       context.TODO(),
					wrapper:   &MockInternalWrapper{},
					Logger:    log.New(&buf, "test-logger", log.LstdFlags),
				})

				wantErr = tt.wantErrMsg != ""
			)

			if tt.before != nil {
				tt.before(n)
			}

			err := n.Close()
			if wantErr {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAMQPNotifier_GetChannel(t *testing.T) {
	d := New(&Config{})
	defer func() { close(d.Channel) }()
	got := d.GetChannel()
	assert.NotNilf(t, got, "GetChannel() = nil, want not nil")
	time.Sleep(10 * time.Millisecond)
}

func TestWebhookNotifier_Notify(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "test:", log.LstdFlags)

		tests = []struct {
			name      string
			channel   chan *model.Notification
			payload   *model.Notification
			wantLog   string
			wantPanic bool
			wantValue bool
		}{
			{
				name:      "nil-channel",
				channel:   nil,
				payload:   nil,
				wantLog:   "channel is nil",
				wantPanic: false,
			},
			{
				name:      "nil-payload",
				channel:   make(chan *model.Notification, 1),
				payload:   nil,
				wantLog:   "payload is nil",
				wantPanic: false,
			},
			{
				name:      "valid-payload",
				channel:   make(chan *model.Notification, 1),
				payload:   &model.Notification{},
				wantLog:   "",
				wantPanic: false,
				wantValue: true,
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("Notify has panicked, expected not to")
				}
			}()

			dn := &AMQPNotifier{
				Config:  &Config{Logger: logger},
				Channel: tt.channel,
			}

			dn.Notify(tt.payload)

			model.CheckLoggerError(&buf, tt.wantLog)

			if tt.wantValue {
				select {
				case value := <-tt.channel:
					assert.Equalf(t, tt.payload, value, "Notify payload = %+v, expected %+v", value, tt.payload)

				default:
					t.Errorf("Notify payload is empty, expected %+v", tt.payload)
				}
			}

			buf.Reset()
		})
	}
}

func TestAMQPNotifier_Run(t *testing.T) {
	var (
		buf     bytes.Buffer
		logger  = log.New(&buf, "test:", log.LstdFlags)
		message = &model.Notification{Data: "test"}

		tests = []struct {
			name   string
			before func(n *AMQPNotifier)
			checks []model.TestCheckNotifierFn
		}{
			{
				name: "success",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return payload, nil
					}

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(nil)
				},
				checks: []model.TestCheckNotifierFn{
					model.CheckLoggerError(&buf, ""),
				},
			},
			{
				name: "fail-jsonMarchal",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return nil, fmt.Errorf("test-jsonMarshal-error")
					}

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(nil)
				},
				checks: []model.TestCheckNotifierFn{
					model.CheckLoggerError(&buf, "test-jsonMarshal-error"),
				},
			},
			{
				name: "fail-Send-timeout",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return payload, nil
					}

					n.wrapper.(*MockInternalWrapper).wait = 50 * time.Millisecond

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(nil)
				},
				checks: []model.TestCheckNotifierFn{
					model.CheckLoggerError(&buf, "message delivery timed out"),
				},
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				n = New(&Config{
					ctx:             context.TODO(),
					wrapper:         &MockInternalWrapper{},
					Logger:          logger,
					DeliveryTimeout: 30,
				})
			)

			if tt.before != nil {
				tt.before(n)
			}

			// start the runner
			go n.Run()
			time.Sleep(10 * time.Millisecond)

			// send a single message and close the channel
			go func() {
				n.Channel <- message
				close(n.Channel)
			}()

			time.Sleep(100 * time.Millisecond)

			for _, c := range tt.checks {
				c(t, n)
			}

			buf.Reset()
		})
	}
}

func TestWebhookNotifier_Deliver(t *testing.T) {
	var (
		buf     bytes.Buffer
		logger  = log.New(&buf, "test:", log.LstdFlags)
		message = &model.Notification{Data: "test"}

		checkSuccess = func(t *testing.T, n model.Notifier, r *model.Result) {
			t.Helper()
			assert.True(t, r.Success)
		}

		tests = []struct {
			name   string
			before func(n *AMQPNotifier)
			checks []model.TestCheckResultFn
		}{
			{
				name: "success",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return payload, nil
					}

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(nil)
				},
				checks: []model.TestCheckResultFn{
					model.CheckResultError(""),
					checkSuccess,
				},
			},
			{
				name: "fail-jsonMarchal",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return nil, fmt.Errorf("test-jsonMarshal-error")
					}

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(nil)
				},
				checks: []model.TestCheckResultFn{
					model.CheckResultError("test-jsonMarshal-error"),
				},
			},
			{
				name: "fail-Send",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return payload, nil
					}

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(fmt.Errorf("test-Send-error"))
				},
				checks: []model.TestCheckResultFn{
					model.CheckResultError("sending message:"),
				},
			},
			{
				name: "fail-Send-timeout",
				before: func(n *AMQPNotifier) {
					var (
						payload = []byte("test")
						w       = n.wrapper.(*MockInternalWrapper)
					)

					n.jsonMarshal = func(v any) ([]byte, error) {
						return payload, nil
					}

					n.wrapper.(*MockInternalWrapper).wait = 50 * time.Millisecond

					w.On("Send", mock.Anything, amqp.NewMessage(payload), (*amqp.SendOptions)(nil)).
						Return(nil)
				},
				checks: []model.TestCheckResultFn{
					model.CheckResultError("message delivery timed out"),
				},
			},
		}
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(&Config{
				ctx:             context.TODO(),
				wrapper:         &MockInternalWrapper{},
				Logger:          logger,
				DeliveryTimeout: 10,
			})

			if tt.before != nil {
				tt.before(n)
			}

			r := n.Deliver(message)
			for _, c := range tt.checks {
				c(t, n, r)
			}

			buf.Reset()
		})
	}
}
