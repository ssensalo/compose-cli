/*
	Copyright (c) 2020 Docker Inc.

	Permission is hereby granted, free of charge, to any person
	obtaining a copy of this software and associated documentation
	files (the "Software"), to deal in the Software without
	restriction, including without limitation the rights to use, copy,
	modify, merge, publish, distribute, sublicense, and/or sell copies
	of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be
	included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED,
	INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
	HOLDERS BE LIABLE FOR ANY CLAIM,
	DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT,
	TORT OR OTHERWISE,
	ARISING FROM, OUT OF OR IN CONNECTION WITH
	THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	"gotest.tools/golden"

	. "github.com/docker/api/tests/framework"
)

type E2eSuite struct {
	Suite
}

func (s *E2eSuite) TestContextHelp() {
	output := s.NewDockerCommand("context", "create", "aci", "--help").ExecOrDie()
	Expect(output).To(ContainSubstring("docker context create aci CONTEXT [flags]"))
	Expect(output).To(ContainSubstring("--location"))
	Expect(output).To(ContainSubstring("--subscription-id"))
	Expect(output).To(ContainSubstring("--resource-group"))
}

func (s *E2eSuite) TestListAndShowDefaultContext() {
	output := s.NewDockerCommand("context", "show").ExecOrDie()
	Expect(output).To(ContainSubstring("default"))
	output = s.NewCommand("docker", "context", "ls").ExecOrDie()
	golden.Assert(s.T(), output, GoldenFile("ls-out-default"))
}

func (s *E2eSuite) TestCreateDockerContextAndListIt() {
	s.NewDockerCommand("context", "create", "test-docker", "--from", "default").ExecOrDie()
	output := s.NewCommand("docker", "context", "ls").ExecOrDie()
	golden.Assert(s.T(), output, GoldenFile("ls-out-test-docker"))
}

func (s *E2eSuite) TestContextListQuiet() {
	s.NewDockerCommand("context", "create", "test-docker", "--from", "default").ExecOrDie()
	output := s.NewCommand("docker", "context", "ls", "-q").ExecOrDie()
	Expect(output).To(Equal(`default
test-docker
`))
}

func (s *E2eSuite) TestInspectDefaultContext() {
	output := s.NewDockerCommand("context", "inspect", "default").ExecOrDie()
	Expect(output).To(ContainSubstring(`"Name": "default"`))
}

func (s *E2eSuite) TestInspectContextNoArgs() {
	output := s.NewDockerCommand("context", "inspect").ExecOrDie()
	Expect(output).To(ContainSubstring(`"Name": "default"`))
}

func (s *E2eSuite) TestInspectContextRegardlessCurrentContext() {
	s.NewDockerCommand("context", "create", "local", "localCtx").ExecOrDie()
	s.NewDockerCommand("context", "use", "localCtx").ExecOrDie()
	output := s.NewDockerCommand("context", "inspect").ExecOrDie()
	Expect(output).To(ContainSubstring(`"Name": "localCtx"`))
}

func (s *E2eSuite) TestContextCreateParseErrorDoesNotDelegateToLegacy() {
	It("should dispay new cli error when parsing context create flags", func() {
		_, err := s.NewDockerCommand("context", "create", "aci", "--subscription-id", "titi").Exec()
		Expect(err.Error()).NotTo(ContainSubstring("unknown flag"))
		Expect(err.Error()).To(ContainSubstring("accepts 1 arg(s), received 0"))
	})
}

func (s *E2eSuite) TestCannotRemoveCurrentContext() {
	s.NewDockerCommand("context", "create", "test-context-rm", "--from", "default").ExecOrDie()
	s.NewDockerCommand("context", "use", "test-context-rm").ExecOrDie()
	_, err := s.NewDockerCommand("context", "rm", "test-context-rm").Exec()
	Expect(err.Error()).To(ContainSubstring("cannot delete current context"))
}

func (s *E2eSuite) TestCanForceRemoveCurrentContext() {
	s.NewDockerCommand("context", "create", "test-context-rmf", "--from", "default").ExecOrDie()
	s.NewDockerCommand("context", "use", "test-context-rmf").ExecOrDie()
	s.NewDockerCommand("context", "rm", "-f", "test-context-rmf").ExecOrDie()
	out := s.NewDockerCommand("context", "ls").ExecOrDie()
	Expect(out).To(ContainSubstring("default *"))
}

func (s *E2eSuite) TestClassicLoginWithparameters() {
	output, err := s.NewDockerCommand("login", "-u", "nouser", "-p", "wrongpasword").Exec()
	Expect(output).To(ContainSubstring("Get https://registry-1.docker.io/v2/: unauthorized: incorrect username or password"))
	Expect(err).NotTo(BeNil())
}

func (s *E2eSuite) TestClassicLoginRegardlessCurrentContext() {
	s.NewDockerCommand("context", "create", "local", "localCtx").ExecOrDie()
	s.NewDockerCommand("context", "use", "localCtx").ExecOrDie()
	output, err := s.NewDockerCommand("login", "-u", "nouser", "-p", "wrongpasword").Exec()
	Expect(output).To(ContainSubstring("Get https://registry-1.docker.io/v2/: unauthorized: incorrect username or password"))
	Expect(err).NotTo(BeNil())
}

func (s *E2eSuite) TestClassicLogin() {
	output, err := s.NewDockerCommand("login", "someregistry.docker.io").Exec()
	Expect(output).To(ContainSubstring("Cannot perform an interactive login from a non TTY device"))
	Expect(err).NotTo(BeNil())
}

func (s *E2eSuite) TestCloudLogin() {
	output, err := s.NewDockerCommand("login", "mycloudbackend").Exec()
	Expect(output).To(ContainSubstring("unknown backend type for cloud login: mycloudbackend"))
	Expect(err).NotTo(BeNil())
}

func (s *E2eSuite) TestSetupError() {
	It("should display an error if cannot shell out to com.docker.cli", func() {
		err := os.Setenv("PATH", s.BinDir)
		Expect(err).To(BeNil())
		err = os.Remove(filepath.Join(s.BinDir, DockerClassicExecutable()))
		Expect(err).To(BeNil())
		output, err := s.NewDockerCommand("ps").Exec()
		Expect(output).To(ContainSubstring("com.docker.cli"))
		Expect(output).To(ContainSubstring("not found"))
		Expect(err).NotTo(BeNil())
	})
}

func (s *E2eSuite) TestLegacy() {
	It("should list all legacy commands", func() {
		output := s.NewDockerCommand("--help").ExecOrDie()
		Expect(output).To(ContainSubstring("swarm"))
	})

	It("should execute legacy commands", func() {
		output, _ := s.NewDockerCommand("swarm", "join").Exec()
		Expect(output).To(ContainSubstring("\"docker swarm join\" requires exactly 1 argument."))
	})

	It("should run local container in less than 10 secs", func() {
		s.NewDockerCommand("pull", "hello-world").ExecOrDie()
		output := s.NewDockerCommand("run", "--rm", "hello-world").WithTimeout(time.NewTimer(20 * time.Second).C).ExecOrDie()
		Expect(output).To(ContainSubstring("Hello from Docker!"))
	})

	It("should execute legacy commands in other moby contexts", func() {
		s.NewDockerCommand("context", "create", "mobyCtx", "--from=default").ExecOrDie()
		s.NewDockerCommand("context", "use", "mobyCtx").ExecOrDie()
		output, _ := s.NewDockerCommand("swarm", "join").Exec()
		Expect(output).To(ContainSubstring("\"docker swarm join\" requires exactly 1 argument."))
	})
}

func (s *E2eSuite) TestLeaveLegacyErrorMessagesUnchanged() {
	output, err := s.NewDockerCommand("foo").Exec()
	golden.Assert(s.T(), output, "unknown-foo-command.golden")
	Expect(err).NotTo(BeNil())
}

func (s *E2eSuite) TestDisplayFriendlyErrorMessageForLegacyCommands() {
	s.NewDockerCommand("context", "create", "example", "test-example").ExecOrDie()
	output, err := s.NewDockerCommand("--context", "test-example", "images").Exec()
	Expect(output).To(Equal("Command \"images\" not available in current context (test-example), you can use the \"default\" context to run this command\n"))
	Expect(err).NotTo(BeNil())
}

func (s *E2eSuite) TestDisplaysAdditionalLineInDockerVersion() {
	output := s.NewDockerCommand("version").ExecOrDie()
	Expect(output).To(ContainSubstring("Azure integration"))
}

func (s *E2eSuite) TestMockBackend() {
	It("creates a new test context to hardcoded example backend", func() {
		s.NewDockerCommand("context", "create", "example", "test-example").ExecOrDie()
		// Expect(output).To(ContainSubstring("test-example context acitest created"))
	})

	It("uses the test context", func() {
		currentContext := s.NewDockerCommand("context", "use", "test-example").ExecOrDie()
		Expect(currentContext).To(ContainSubstring("test-example"))
		output := s.NewDockerCommand("context", "ls").ExecOrDie()
		golden.Assert(s.T(), output, GoldenFile("ls-out-test-example"))
		output = s.NewDockerCommand("context", "show").ExecOrDie()
		Expect(output).To(ContainSubstring("test-example"))
	})

	It("can run ps command", func() {
		output := s.NewDockerCommand("ps").ExecOrDie()
		lines := Lines(output)
		Expect(len(lines)).To(Equal(3))
		Expect(lines[2]).To(ContainSubstring("1234                alpine"))
	})

	It("can run quiet ps command", func() {
		output := s.NewDockerCommand("ps", "-q").ExecOrDie()
		lines := Lines(output)
		Expect(len(lines)).To(Equal(2))
		Expect(lines[0]).To(Equal("id"))
		Expect(lines[1]).To(Equal("1234"))
	})

	It("can run ps command with all ", func() {
		output := s.NewDockerCommand("ps", "-q", "--all").ExecOrDie()
		lines := Lines(output)
		Expect(len(lines)).To(Equal(3))
		Expect(lines[0]).To(Equal("id"))
		Expect(lines[1]).To(Equal("1234"))
		Expect(lines[2]).To(Equal("stopped"))
	})

	It("can run inspect command on container", func() {
		golden.Assert(s.T(), s.NewDockerCommand("inspect", "id").ExecOrDie(), "inspect-id.golden")
	})

	It("can run 'run' command", func() {
		output := s.NewDockerCommand("run", "nginx", "-p", "80:80").ExecOrDie()
		Expect(output).To(ContainSubstring("Running container \"nginx\" with name"))
	})
}

func TestE2e(t *testing.T) {
	suite.Run(t, new(E2eSuite))
}