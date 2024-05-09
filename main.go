package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

// https://github.com/urfave/cli/issues/731
func run(args []string, outputDir string, client Client) error {
	config, err := load(outputDir)
	if err != nil {
		return err
	}
	app := &cli.App{
		Name:  "cgpt",
		Usage: "Chat with GPT",
		Commands: []*cli.Command{
			{
				Name:  "reset",
				Usage: "Reset session",
				Action: func(ctx *cli.Context) error {
					sessionId, err := uuid.NewRandom()
					if err != nil {
						return err
					}

					// configのsession更新
					config.SessionId = sessionId.String()
					err = config.update()
					if err != nil {
						return nil
					}
					return nil
				},
			},
			{
				Name:  "model",
				Usage: "Change GPT model",
				Action: func(ctx *cli.Context) error {
					prompt := promptui.Select{
						Label: "Select Model",
						Items: []string{openai.GPT3Dot5Turbo, openai.GPT4, openai.GPT4Turbo},
					}

					_, model, err := prompt.Run()

					if err != nil {
						return err
					}

					// configのModel更新
					config.Model = model
					err = config.update()
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "Show Config or Session",
				Subcommands: []*cli.Command{
					{
						Name:  "config",
						Usage: "Show config",
						Action: func(ctx *cli.Context) error {
							jsonData, err := json.MarshalIndent(config, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(jsonData))
							return nil
						},
					},
					{
						Name:  "session",
						Usage: "Show session",
						Action: func(ctx *cli.Context) error {
							session := Session{Dir: filepath.Join(outputDir, "session")}
							err := session.read(config.SessionId)
							if err != nil {
								return err
							}
							jsonData, err := json.MarshalIndent(session, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(jsonData))

							return nil
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "edit",
				Usage:   "Edit message in VS Code",
				Aliases: []string{"e"},
			},
		},
		Action: func(ctx *cli.Context) error {
			session := Session{Dir: filepath.Join(outputDir, "session")}

			convesationId := config.SessionId
			sessionPath := filepath.Join(session.Dir, fmt.Sprintf(`%s.json`, convesationId))

			// 会話履歴ファイルの作成or読み取り
			if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
				session.Messages = []openai.ChatCompletionMessage{}
				session.create(convesationId)
			} else {
				err = session.read(convesationId)
				if err != nil {
					return err
				}
			}

			var content *string
			if ctx.Bool("edit") {
				editor := VSCode{Dir: filepath.Join(outputDir, "message")}
				var err error
				content, err = editor.save()
				if err != nil {
					return err
				}
			} else {
				contentStr := ctx.Args().Get(0)
				content = &contentStr
			}

			session.Messages = append(session.Messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: *content})
			result, err := client.runOpenAI(config.Model, session.Messages)
			if err != nil {
				return err
			}
			session.Messages = append(session.Messages, *result)

			// sessionFileを更新
			err = session.update(convesationId)
			if err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(args); err != nil {
		return err
	}
	return nil
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	outputDir := filepath.Join(homeDir, ".cgpt")

	client := OpenAIClient{Client: openai.NewClient(os.Getenv("OPENAI_API_KEY"))}

	if err = run(os.Args, outputDir, client); err != nil {
		log.Fatal(err)
	}
}

func load(outputDir string) (*Config, error) {
	createDirIfNotExists(filepath.Join(outputDir, "session"))
	createDirIfNotExists(filepath.Join(outputDir, "message"))

	config := Config{Path: filepath.Join(outputDir, "config.json")}
	// config.json作成or読み取り
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		sessionId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		config.SessionId = sessionId.String()
		config.Model = openai.GPT3Dot5Turbo
		err = config.create()
		if err != nil {
			return nil, err
		}
	} else {
		err = config.read()
		if err != nil {
			return nil, err
		}
	}
	return &config, nil
}

func createDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
