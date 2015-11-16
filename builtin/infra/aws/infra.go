package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/otto/helper/bindata"
	"github.com/hashicorp/otto/helper/sshagent"
	//"github.com/hashicorp/otto/helper/terraform"
	"github.com/hashicorp/otto/ui"
	"github.com/kuuyee/otto-learn/helper/terraform"
	"github.com/kuuyee/otto-learn/infrastructure"
	"github.com/mitchellh/go-homedir"
)

// go:generate go-bindata -pkg=aws -nomemcopy -nometadata ./data/...

// Infra返回一个infrastructure.Infrastructure实现
// 这是一个工厂函数
func Infra() (infrastructure.Infrastructure, error) {
	return &terraform.Infrastructure{
		CredsFunc:       creds,
		VerifyCredsFunc: verifyCreds,
		Bindata: &bindata.Data{
			Asset:    Asset,
			AssetDir: AssetDir,
		},
		Variables: map[string]string{
			"aws_region": "us-east-1",
		},
	}, nil
}

func creds(ctx *infrastructure.Context) (map[string]string, error) {
	fields := []*ui.InputOpts{
		&ui.InputOpts{
			Id:          "aws_access_key",
			Query:       "AWS Access Key",
			Description: "AWS access key used for API calls",
			EnvVars:     []string{"AWS_ACCESS_KEY_ID"},
		},
		&ui.InputOpts{
			Id:          "aws_secret_key",
			Query:       "AWS Secret Key",
			Description: "AWS secret key used for API calls",
			EnvVars:     []string{"AWS_SECRET_KEY_ID"},
		},
		&ui.InputOpts{
			Id:          "ssh_public_key_path",
			Query:       "SSH Public Key Path",
			Description: "Path to an SSH public key that will be granted access to EC2 instances",
			Default:     "~/.ssh/id_rsa.pub",
			EnvVars:     []string{"AWS_SSH_PUBLIC_KEY_PATH"},
		},
	}

	result := make(map[string]string, len(fields))
	for _, f := range fields {
		value, err := ctx.Ui.Input(f)
		if err != nil {
			return nil, err
		}

		result[f.Id] = value
	}

	// 获得SSH public key内容
	sshPath, err := homedir.Expand(result["ssh_public_key_path"])
	if err != nil {
		return nil, fmt.Errorf("展开SSH key的homedir报错：%s", err)
	}

	sshKey, err := ioutil.ReadFile(sshPath)
	if err != nil {
		return nil, fmt.Errorf("读取SSH key报错：%s", err)
	}
	result["ssh_public_key"] = string(sshKey)

	return result, nil
}

func verifyCreds(ctx *infrastructure.Context) error {
	found, err := sshagent.HasKey(ctx.InfraCreds["ssh_public_key"])
	if err != nil {
		return sshAgentError(err)
	}
	if !found {
		ok, _ := guestAndLoadPrivateKey(ctx.Ui, ctx.InfraCreds["ssh_public_key_path"])
		if ok {
			ctx.Ui.Message(
				"发现一个private key并装载。Otto将会检查\n" +
					"SSH Agent如果正确的key被装载")
			found, err = sshagent.HasKey(ctx.InfraCreds["ssh_public_key"])
			if err != nil {
				return sshAgentError(err)
			}
		}
	}

	if !found {
		return sshAgentError(fmt.Errorf(
			"You specified an SSH public key of: %q, but the private key from this\n"+
				"keypair is not loaded the SSH Agent. To load it, run:\n\n"+
				"  ssh-add [PATH_TO_PRIVATE_KEY]",
			ctx.InfraCreds["ssh_public_key_path"]))
	}
	return nil
}

func sshAgentError(err error) error {
	return fmt.Errorf(
		"Otto uses your SSH Agent to authenticate with instances created in\n"+
			"AWS, but it could not verify that your SSH key is loaded into the agent.\n"+
			"The error message follows:\n\n%s", err)
}

// guestAndLoadPrivateKey 获得private key路径，只是依赖判断去除.pub部分的内容
// 如果存在则直接装在agent
func guestAndLoadPrivateKey(ui ui.Ui, pubKeyPath string) (bool, error) {
	fullPath, err := homedir.Expand(pubKeyPath)
	if err != nil {
		return false, err
	}
	if !strings.HasSuffix(fullPath, ".pub") {
		return false, fmt.Errorf("没有找到.pub后缀的文件")
	}
	privKeyGuess := strings.TrimSuffix(fullPath, ".pub")
	if _, err := os.Stat(privKeyGuess); os.IsNotExist(err) {
		return false, fmt.Errorf("文件不存在!")
	}

	ui.Header("装载key到SSH Agent")
	ui.Message(fmt.Sprintf(
		"在SSH Agent中没有你提供的key (%s)", pubKeyPath))
	ui.Message(fmt.Sprintf(
		"然而，Otto在这里：%s 找的了一个private key", privKeyGuess))
	ui.Message(fmt.Sprintf(
		"自动运行'ssh-add %s'.", privKeyGuess))
	ui.Message("如果你的SSH key有密码，你会看到如下提示")
	ui.Message("")

	if err := sshagent.Add(ui, privKeyGuess); err != nil {
		return false, err
	}

	return true, nil
}
