package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	host string
	port int
	pool *ClientPool
)

func main() {
	rootCmd := newRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 添加一些测试数据
	client, err := pool.Get()
	if err != nil {
		fmt.Printf("Failed to get client: %v\n", err)
		return
	}
	defer pool.Put(client)

	client.Do("SET", "test1", "value1")
	client.Do("SET", "test2", "value2")
	client.Do("SET", "other", "value3")

	// 测试 KEYS 命令
	keys, err := client.Do("KEYS", "test*")
	if err != nil {
		fmt.Printf("Error executing KEYS command: %v\n", err)
	} else {
		fmt.Printf("Keys matching 'test*': %v\n", keys)
	}
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "literedis-cli",
		Short: "LiteRedis CLI - A command line interface for LiteRedis",
		Run:   runInteractiveMode,
	}

	rootCmd.PersistentFlags().StringVarP(&host, "host", "h", "localhost", "Server hostname")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 6379, "Server port")

	rootCmd.AddCommand(CreateCommands(&host, &port)...)
	rootCmd.AddCommand(newCompletionCommand())

	pool = NewClientPool(host, port, 10) // 创建一个大小为10的连接池

	return rootCmd
}

func runInteractiveMode(cmd *cobra.Command, args []string) {
	client, err := pool.Get()
	if err != nil {
		fmt.Printf("Failed to get client: %v\n", err)
		return
	}
	defer pool.Put(client)

	repl := NewREPL(client)
	repl.Run()
}

func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(literedis-cli completion bash)

# To load completions for each session, execute once:
Linux:
  $ literedis-cli completion bash > /etc/bash_completion.d/literedis-cli
MacOS:
  $ literedis-cli completion bash > /usr/local/etc/bash_completion.d/literedis-cli

Zsh:

$ source <(literedis-cli completion zsh)

# To load completions for each session, execute once:
$ literedis-cli completion zsh > "${fpath[1]}/_literedis-cli"

Fish:

$ literedis-cli completion fish | source

# To load completions for each session, execute once:
$ literedis-cli completion fish > ~/.config/fish/completions/literedis-cli.fish

PowerShell:

PS> literedis-cli completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> literedis-cli completion powershell > literedis-cli.ps1
# and source this file from your PowerShell profile.
`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
}
