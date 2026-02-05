package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/saurabh-git-dev/dokploy-cli/dokploy"

	cli "github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = ""
)

func main() {
	app := &cli.App{
		Name:  "dokploy cli",
		Usage: "Manage Dokploy projects, compose apps, and domains",
		Version: func() string {
			if commit == "" {
				return version
			}
			return fmt.Sprintf("%s (%s)", version, commit)
		}(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Usage:    "Dokploy base API URL (or set DOKPLOY_URL)",
				EnvVars:  []string{"DOKPLOY_URL"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "key",
				Usage:    "Dokploy API key (or set DOKPLOY_API_KEY)",
				EnvVars:  []string{"DOKPLOY_API_KEY"},
				Required: true,
			},
		},
		Commands: []*cli.Command{
			projectCommand(),
			composeCommand(),
			domainCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newClientFromCtx(c *cli.Context) (*dokploy.Client, error) {
	url := c.String("url")
	key := c.String("key")
	return dokploy.NewClient(url, key)
}

// PROJECT COMMANDS

func projectCommand() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage projects",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create/Get a project and its default environment",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Usage: "Project name", Required: true},
					&cli.StringFlag{Name: "description", Usage: "Project description"},
					&cli.StringFlag{Name: "environment", Usage: "Environment name", Value: "production"},
					&cli.StringFlag{Name: "return", Usage: "Which ID to print (projectId/environmentId/both)", Value: "environmentId"},
				},
				Action: func(c *cli.Context) error {
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}

					var projectId, envId string
					projectId, envId, err = dokploy.GetProject(c.Context, client, c.String("name"), c.String("environment"))
					if projectId == "" {
						projectId, envId, err = dokploy.CreateProject(c.Context, client, c.String("name"), c.String("description"), c.String("environment"))
						if err != nil {
							return err
						}
					}

					mode := strings.ToLower(c.String("return"))
					switch mode {
					case "", "environmentid":
						fmt.Println(envId)
					case "projectid":
						fmt.Println(projectId)
					case "both":
						fmt.Printf("%s %s\n", projectId, envId)
					default:
						return fmt.Errorf("invalid --return value %q, must be one of: projectId, environmentId, both", mode)
					}
					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Project ID", Required: true},
				},
				Action: func(c *cli.Context) error {
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}
					id := c.String("id")
					if err := dokploy.DeleteProject(c.Context, client, id); err != nil {
						return err
					}
					fmt.Println("Deleted project", id)
					return nil
				},
			},
		},
	}
}

// COMPOSE COMMANDS

func composeCommand() *cli.Command {
	return &cli.Command{
		Name:  "compose",
		Usage: "Manage compose apps",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Get a compose app by ID or name",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Compose ID"},
					&cli.StringFlag{Name: "name", Usage: "Compose name"},
				},
				Action: func(c *cli.Context) error {
					if c.String("id") == "" && c.String("name") == "" {
						return errors.New("either --id or --name is required")
					}
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}
					out, err := dokploy.GetCompose(c.Context, client, c.String("id"), c.String("name"))
					if err != nil {
						return err
					}
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(out)
				},
			},
			{
				Name:  "create",
				Usage: "Create or update a compose app",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Compose ID (for update)"},
					&cli.StringFlag{Name: "name", Usage: "Compose name"},
					&cli.StringFlag{Name: "environmentId", Usage: "Environment ID", Required: true},
					&cli.StringFlag{Name: "compose-file", Usage: "Path to docker compose file", Required: true, TakesFile: true},
					&cli.StringSliceFlag{Name: "env-vars", Usage: "Environment variables in KEY=VALUE form (repeatable)"},
				},
				Action: func(c *cli.Context) error {
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}

					composePath := c.String("compose-file")
					content, err := os.ReadFile(composePath)
					if err != nil {
						return err
					}

					envMap := map[string]string{}
					for _, kv := range c.StringSlice("env-vars") {
						parts := strings.SplitN(kv, "=", 2)
						if len(parts) != 2 {
							return fmt.Errorf("invalid env var %q, expected KEY=VALUE", kv)
						}
						envMap[parts[0]] = parts[1]
					}
					id, err := dokploy.CreateOrUpdateCompose(
						c.Context,
						client,
						c.String("id"),
						c.String("name"),
						c.String("environmentId"),
						string(content),
						envMap,
					)
					if err != nil {
						return err
					}
					fmt.Println(id)
					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a compose app",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Compose ID", Required: true},
					&cli.BoolFlag{Name: "delete-volumes", Usage: "Also delete associated volumes (default: true)", Value: true},
				},
				Action: func(c *cli.Context) error {
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}
					id := c.String("id")
					deleteVolumes := c.Bool("delete-volumes")
					if err := dokploy.DeleteCompose(c.Context, client, id, deleteVolumes); err != nil {
						return err
					}
					fmt.Println("Deleted compose", id)
					return nil
				},
			},
			{
				Name:  "deploy",
				Usage: "Deploy a compose app",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Compose ID", Required: true},
				},
				Action: func(c *cli.Context) error {
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}
					id := c.String("id")
					if err := dokploy.DeployCompose(c.Context, client, id); err != nil {
						return err
					}
					fmt.Println("Deployed compose", id)
					return nil
				},
			},
		},
	}
}

// DOMAIN COMMANDS

func domainCommand() *cli.Command {
	return &cli.Command{
		Name:  "domain",
		Usage: "Manage domains",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create or update a domain",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Domain ID (for update)"},
					&cli.StringFlag{Name: "host", Usage: "Domain host", Required: true},
					&cli.StringFlag{Name: "path", Usage: "Path", Value: "/"},
					&cli.IntFlag{Name: "port", Usage: "Service port", Required: true},
					&cli.StringFlag{Name: "serviceName", Usage: "Service name", Required: true},
					&cli.StringFlag{Name: "composeId", Usage: "Compose ID", Required: true},
					&cli.StringFlag{Name: "certificateType", Usage: "Certificate type (none/letsencrypt)", Value: "none"},
					&cli.BoolFlag{Name: "https", Usage: "Enable HTTPS"},
				},
				Action: func(c *cli.Context) error {
					certType := strings.ToLower(c.String("certificateType"))
					switch certType {
					case "none", "letsencrypt":
						// ok
					default:
						return fmt.Errorf("invalid certificateType %q, must be one of: none, letsencrypt", certType)
					}

					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}

					id, err := dokploy.CreateOrUpdateDomain(
						c.Context,
						client,
						c.String("id"),
						c.String("host"),
						c.String("path"),
						c.Int("port"),
						c.String("serviceName"),
						c.String("composeId"),
						certType,
						c.Bool("https"),
					)
					if err != nil {
						return err
					}
					fmt.Println(id)
					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a domain",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Domain ID", Required: true},
				},
				Action: func(c *cli.Context) error {
					client, err := newClientFromCtx(c)
					if err != nil {
						return err
					}
					id := c.String("id")
					if err := dokploy.DeleteDomain(c.Context, client, id); err != nil {
						return err
					}
					fmt.Println("Deleted domain", id)
					return nil
				},
			},
		},
	}
}
