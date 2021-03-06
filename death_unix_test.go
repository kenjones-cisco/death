// +build linux bsd darwin

package death

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
)

type Unhashable map[string]interface{}

func (u Unhashable) Close() error {
	return nil
}

func TestDeath(t *testing.T) {
	defer seelog.Flush()

	Convey("Validate death handles unhashable types", t, func() {
		u := make(Unhashable)
		death := NewDeath(syscall.SIGTERM)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
		death.WaitForDeath(u)
	})

	Convey("Validate death happens cleanly", t, func() {
		death := NewDeath(syscall.SIGTERM)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
		death.WaitForDeath()

	})

	Convey("Validate death happens with other signals", t, func() {
		death := NewDeath(syscall.SIGHUP)
		closeMe := &CloseMe{}
		syscall.Kill(os.Getpid(), syscall.SIGHUP)
		death.WaitForDeath(closeMe)
		So(closeMe.Closed, ShouldEqual, 1)
	})

	Convey("Validate death happens with a manual call", t, func() {
		death := NewDeath(syscall.SIGHUP)
		closeMe := &CloseMe{}
		death.FallOnSword()
		death.WaitForDeath(closeMe)
		So(closeMe.Closed, ShouldEqual, 1)
	})

	Convey("Validate multiple sword falls do not block", t, func() {
		death := NewDeath(syscall.SIGHUP)
		closeMe := &CloseMe{}
		death.FallOnSword()
		death.FallOnSword()
		death.WaitForDeath(closeMe)
		So(closeMe.Closed, ShouldEqual, 1)
	})

	Convey("Validate multiple sword falls do not block even after we have exited waitForDeath", t, func() {
		death := NewDeath(syscall.SIGHUP)
		closeMe := &CloseMe{}
		death.FallOnSword()
		death.FallOnSword()
		death.WaitForDeath(closeMe)
		death.FallOnSword()
		death.FallOnSword()
		So(closeMe.Closed, ShouldEqual, 1)
	})

	Convey("Validate death gives up after timeout", t, func() {
		death := NewDeath(syscall.SIGHUP)
		death.SetTimeout(10 * time.Millisecond)
		neverClose := &neverClose{}
		syscall.Kill(os.Getpid(), syscall.SIGHUP)
		death.WaitForDeath(neverClose)

	})

	Convey("Validate death uses new logger", t, func() {
		death := NewDeath(syscall.SIGHUP)
		closeMe := &CloseMe{}
		logger := &MockLogger{}
		death.SetLogger(logger)

		syscall.Kill(os.Getpid(), syscall.SIGHUP)
		death.WaitForDeath(closeMe)
		So(closeMe.Closed, ShouldEqual, 1)
		So(logger.Logs, ShouldNotBeEmpty)
	})

	Convey("Close multiple things with one that fails the timer", t, func() {
		death := NewDeath(syscall.SIGHUP)
		death.SetTimeout(10 * time.Millisecond)
		neverClose := &neverClose{}
		closeMe := &CloseMe{}
		syscall.Kill(os.Getpid(), syscall.SIGHUP)
		death.WaitForDeath(neverClose, closeMe)
		So(closeMe.Closed, ShouldEqual, 1)
	})

	Convey("Close with anonymous function", t, func() {
		death := NewDeath(syscall.SIGHUP)
		death.SetTimeout(5 * time.Millisecond)
		closeMe := &CloseMe{}
		syscall.Kill(os.Getpid(), syscall.SIGHUP)
		death.WaitForDeathWithFunc(func() {
			closeMe.Close()
			So(true, ShouldBeTrue)
		})
		So(closeMe.Closed, ShouldEqual, 1)
	})

}

type MockLogger struct {
	Logs []interface{}
}

func (l *MockLogger) Info(v ...interface{}) {
	for _, log := range v {
		l.Logs = append(l.Logs, log)
	}
}

func (l *MockLogger) Debug(v ...interface{}) {
	for _, log := range v {
		l.Logs = append(l.Logs, log)
	}
}

func (l *MockLogger) Error(v ...interface{}) error {
	for _, log := range v {
		l.Logs = append(l.Logs, log)
	}
	return nil
}

func (l *MockLogger) Warn(v ...interface{}) error {
	for _, log := range v {
		l.Logs = append(l.Logs, log)
	}
	return nil
}

type neverClose struct {
}

func (n *neverClose) Close() error {
	time.Sleep(2 * time.Minute)
	return nil
}

type CloseMe struct {
	Closed int
}

func (c *CloseMe) Close() error {
	c.Closed++
	return errors.New("I have been closed")
}
