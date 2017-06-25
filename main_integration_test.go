package main_test

import (
	// TODO: Add this back when some tests are written:
	// . "bitbucket.org/mexisme/get-secrets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bitbucket.org/mexisme/get-secrets/dotenv"
	s3ish "bitbucket.org/mexisme/get-secrets/files/s3"
	urlish "bitbucket.org/mexisme/get-secrets/files/s3/s3url"
	"fmt"
	"strings"

	"github.com/subosito/gotenv"
)

// EnvRead TODO
type EnvRead struct {
	bodies []string
	envs   []gotenv.Env
}

func (s *EnvRead) readCallback(path, body string) error {
	if path == "" {
		return fmt.Errorf("Empty path")
	}
	if body == "" {
		return fmt.Errorf("Empty body")
	}

	env := gotenv.Parse(strings.NewReader(body))
	s.bodies = append(s.bodies, body)
	s.envs = append(s.envs, env)

	return nil
}

var _ = Describe("The main Integration Tests", func() {
	var (
		env   *EnvRead
		s3url *urlish.Path
		s3    *s3ish.Details
	)

	BeforeEach(func() {
		env = &EnvRead{
			bodies: make([]string, 0),
			envs:   make([]gotenv.Env, 0),
		}

		// s3url = urlish.New().WithURL(viper.GetString("s3.path"))
		s3url = urlish.New().WithURL("s3://kiwiops-ecs-staging-env/stuff-brightcove-video-service")
		s3 = s3ish.New().WithSource(s3url)
	})

	Describe("when reading from S3", func() {
		var (
			s3lists []string
			err     error
		)

		BeforeEach(func() {
			s3lists, err = s3.List()
		})

		It("reads the file-list from S3", func() {
			Expect(s3lists).To(Not(BeEmpty()))
			Expect(err).To(BeNil())
		})

		Describe("(including object contents)", func() {
			It("reads the env files from S3", func() {
				errs := s3.ReadList(s3lists, env.readCallback)

				Expect(env.envs).To(Not(BeEmpty()))
				Expect(errs).To(BeNil())
			})

			It("fails to read the env files from S3", func() {
				s3lists = append(s3lists, "")
				s3lists = append(s3lists, "lol")
				errs := s3.ReadList(s3lists, env.readCallback)

				Expect(errs).To(Not(BeNil()))
			})
		})

		Describe("(including parsing the object contents)", func() {
			var envs *dotenv.DotEnvs

			BeforeEach(func() {
				envs = dotenv.New()
			})

			It("reads the env files from S3", func() {
				errs := s3.ReadList(s3lists, envs.AddFromString)

				Expect(errs).To(BeNil())
			})
		})
	})
})