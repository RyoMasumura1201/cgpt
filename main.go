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
				Usage: "Reset chat",
				Action: func(ctx *cli.Context) error {
					chatId, err := uuid.NewRandom()
					if err != nil {
						return err
					}

					// configのchat更新
					config.ChatId = chatId.String()
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
						Items: []string{
							openai.GPT4o,
							openai.GPT3Dot5Turbo,
							openai.GPT4,
							openai.GPT4Turbo,
						},
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
				Usage: "Show Config or Chat",
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
						Name:  "chat",
						Usage: "Show chat",
						Action: func(ctx *cli.Context) error {
							chat := Chat{Dir: filepath.Join(outputDir, "chat")}
							err := chat.read(config.ChatId)
							if err != nil {
								return err
							}
							jsonData, err := json.MarshalIndent(chat, "", "  ")
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
			chat := Chat{Dir: filepath.Join(outputDir, "chat")}

			convesationId := config.ChatId
			chatPath := filepath.Join(chat.Dir, fmt.Sprintf(`%s.json`, convesationId))

			// 会話履歴ファイルの作成or読み取り
			if _, err := os.Stat(chatPath); os.IsNotExist(err) {
				chat.Messages = []openai.ChatCompletionMessage{}
				chat.create(convesationId)
			} else {
				err = chat.read(convesationId)
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

			chat.Messages = append(chat.Messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: *content})
			result, err := client.runOpenAI(config.Model, chat.Messages)
			if err != nil {
				return err
			}
			chat.Messages = append(chat.Messages, *result)

			// chatFileを更新
			err = chat.update(convesationId)
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
	createDirIfNotExists(filepath.Join(outputDir, "chat"))
	createDirIfNotExists(filepath.Join(outputDir, "message"))

	config := Config{Path: filepath.Join(outputDir, "config.json")}
	// config.json作成or読み取り
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		chatId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		config.ChatId = chatId.String()
		config.Model = openai.GPT4o
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
