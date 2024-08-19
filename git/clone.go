package git

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitOptions struct {
	Enabled  bool   `form:"enabled" json:"enabled"`
	Repo     string `form:"repo" json:"repo"`
	AppDir   string
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

func Pull(opts GitOptions) error {
	// 初始化一个指向本地目录的新仓库（如果不存在则创建）
	r, err := git.PlainClone(opts.AppDir, false, &git.CloneOptions{
		URL:      opts.Repo,
		Progress: os.Stdout,
		Auth: &http.BasicAuth{
			Username: opts.Username,
			Password: opts.Password, // 使用PAT代替密码
		},
	})
	if err != nil {
		return err
	}

	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	// 执行 git pull
	pullOpts := &git.PullOptions{
		RemoteName: "origin", // 默认的远程仓库名称
		Auth: &http.BasicAuth{
			Username: opts.Username,
			Password: opts.Password,
		},
		Progress: os.Stdout,
	}

	err = worktree.Pull(pullOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("Pull failed: %v", err)
	}
	fmt.Println("Pull successful.")
	return nil
}
