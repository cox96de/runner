//go:build linux

package db

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"

	"github.com/cockroachdb/errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/cox96de/runner/db"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGenDDL(t *testing.T) {
	if os.Getenv("GEN_DDL") != "true" {
		t.Skipf("skip gen ddl")
	}
	t.Run("GenPG", func(t *testing.T) {
		name := "runner-gen-ddl-postgres"
		_, err := run("docker", "run", "--name", name, "--rm", "-e", "POSTGRES_PASSWORD=123456", "-d", "postgres")
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = run("docker", "stop", name)
		})
		ip, err := getContainerIP(name)
		require.NoError(t, err)

		dsn := fmt.Sprintf("postgresql://postgres:%s@%s:5432/postgres", "123456", ip)
		var sqls []string
		err = backoff.Retry(func() error {
			conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				return err
			}
			dbCli := db.NewClient(conn)
			sqls, err = dbCli.ToMigrateSQL()
			return err
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 10))
		require.NoError(t, err)
		t.Logf("%+v", sqls)
		err = os.WriteFile("ddl.postgres.sql", []byte(strings.Join(sqls, ";\n")), 0o644)
		require.NoError(t, err)
	})
	t.Run("GenMySQL", func(t *testing.T) {
		name := "runner-gen-ddl-mysql"
		_, err := run("docker", "run", "--name", name, "--rm", "-e", "MYSQL_ROOT_PASSWORD=123456", "-e", "MYSQL_DATABASE=db", "-d", "mysql")
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = run("docker", "stop", name)
		})
		ip, err := getContainerIP(name)
		require.NoError(t, err)
		ip = strings.TrimPrefix(strings.TrimSpace(ip), "'")
		ip = strings.TrimSuffix(strings.TrimSpace(ip), "'")
		dsn := fmt.Sprintf("root:%s@tcp(%s:3306)/db", "123456", ip)
		var sqls []string
		err = backoff.Retry(func() error {
			conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				return err
			}
			dbCli := db.NewClient(conn)
			sqls, err = dbCli.ToMigrateSQL()
			return err
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 10))
		require.NoError(t, err)
		t.Logf("%+v", sqls)
		err = os.WriteFile("ddl.mysql.sql", []byte(strings.Join(sqls, ";\n")), 0o644)
		require.NoError(t, err)
	})
	t.Run("GenSQLite", func(t *testing.T) {
		var sqls []string
		err := backoff.Retry(func() error {
			conn, err := gorm.Open(sqlite.Open("file:runner-gen-ddl-sqlite?mode=memory"), &gorm.Config{})
			if err != nil {
				return err
			}
			dbCli := db.NewClient(conn)
			sqls, err = dbCli.ToMigrateSQL()
			return err
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 10))
		require.NoError(t, err)
		t.Logf("%+v", sqls)
		err = os.WriteFile("ddl.sqlite.sql", []byte(strings.Join(sqls, ";\n")), 0o644)
		require.NoError(t, err)
	})
}

func getContainerIP(containerName string) (string, error) {
	output, err := run("docker", "inspect", "-f",
		"'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'", containerName)
	if err != nil {
		return "", err
	}
	ip := strings.TrimSpace(output)
	ip = strings.TrimPrefix(ip, "'")
	ip = strings.TrimSuffix(ip, "'")
	return ip, nil
}

func run(command string, args ...string) (output string, err error) {
	cmd := exec.Command(command, args...)
	ouptut, err := cmd.CombinedOutput()
	if err != nil {
		return string(ouptut), errors.WithMessage(err, "failed to run command")
	}
	if exitCode := cmd.ProcessState.ExitCode(); exitCode != 0 {
		return string(ouptut), errors.Errorf("command exited with code %d", exitCode)
	}
	return string(ouptut), nil
}
