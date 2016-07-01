package main

import "io/ioutil"
import "os/exec"
import "encoding/json"
import "net/http"
import "net/url"
import "strings"
import "os"
import "path/filepath"

type versionBotClient struct {
	identifier  string
	versionName string
	notation    string

	version string
}

func newVersionBotClient(identifier string, versionName string, notation string) *versionBotClient {
	c := new(versionBotClient)
	c.identifier = identifier
	c.versionName = versionName
	c.notation = notation

	return c
}

func (c *versionBotClient) bumpVersion() {
	resp, err := http.PostForm("http://version-bot.herokuapp.com/v1/versions", url.Values{
		"identifier":    []string{c.identifier},
		"short_version": []string{c.versionName},
	})
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	c.version = response[c.notation].(string)
}

func (c *versionBotClient) useGitVersion() error {
	cmd := "git"
	args := []string{"rev-parse", "--short", "HEAD"}

	cmdOut, err := exec.Command(cmd, args...).Output()
	if err != nil {
		c.version = c.versionName
		return err
	}

	gitHex := strings.TrimSpace(string(cmdOut))

	c.version = c.versionName + "." + gitHex

	return nil
}

func (c *versionBotClient) createGitTag() error {
	cmd := "git"
	args := []string{"tag", "v" + c.version}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		return err
	}

	return nil
}

func (c *versionBotClient) pushGitTag() error {
	cmd := "git"
	args := []string{"push", "origin", "v" + c.version}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		return err
	}

	return nil
}

func (c *versionBotClient) writeHeader(headerPath string) error {
	os.MkdirAll(filepath.Dir(headerPath), 0777)

	f, err := os.Create(headerPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("#define BUILD_NUMBER " + c.version)
	if err != nil {
		return err
	}

	return nil
}
