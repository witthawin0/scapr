package new

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/witthawin0/scapr/config"
	"github.com/witthawin0/scapr/internal/pkg/helper"
)

type Project struct {
	ProjectName string `survey:"name"`
	Framework   string
	Database    string
	ORM         string
}

var CmdNew = &cobra.Command{
	Use:     "new",
	Example: "scapr new demo-api",
	Short:   "Create a new project with scapr template.",
	Long:    "Create a new project with a customizable scapr template.",
	Run:     run,
}

var (
	repoURL string
)

func init() {
	CmdNew.Flags().StringVarP(&repoURL, "repo-url", "r", repoURL, "layout repository URL")
}

func NewProject() *Project {
	return &Project{}
}

func run(cmd *cobra.Command, args []string) {
	p := NewProject()

	// Check if a project name is provided as an argument
	if len(args) == 0 {
		// Prompt for the project name if not provided
		err := survey.AskOne(&survey.Input{
			Message: "What is your project name?",
			Help:    "Enter the name of your project.",
		}, &p.ProjectName, survey.WithValidator(survey.Required))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	} else {
		// Use the first argument as the project name
		p.ProjectName = args[0]
	}

	// Gather user preferences
	frameworkPrompt := &survey.Select{
		Message: "Select your HTTP framework:",
		Options: []string{"Gin", "Fiber", "Echo"},
	}
	_ = survey.AskOne(frameworkPrompt, &p.Framework)

	databasePrompt := &survey.Select{
		Message: "Select your database:",
		Options: []string{"PostgreSQL", "MySQL", "MongoDB"},
	}
	_ = survey.AskOne(databasePrompt, &p.Database)

	ormPrompt := &survey.Select{
		Message: "Select your ORM/SQL mapper:",
		Options: []string{"GORM", "SQLC", "None"},
	}
	_ = survey.AskOne(ormPrompt, &p.ORM)

	// Clone template
	yes, err := p.cloneTemplate()
	if err != nil || !yes {
		return
	}

	// Replace package name and perform initial setup
	if err := p.replacePackageName(); err != nil {
		return
	}
	if err := p.configureProject(); err != nil {
		return
	}
	if err := p.modTidy(); err != nil {
		return
	}

	p.rmGit()
	fmt.Printf("\n🎉 Project \u001B[36m%s\u001B[0m created successfully!\n", p.ProjectName)
	fmt.Printf("\nDone. Now run:\n\n")
	fmt.Printf("› \033[36mcd %s \033[0m\n", p.ProjectName)
	fmt.Printf("› \033[36mnunu run \033[0m\n\n")
}

func (p *Project) cloneTemplate() (bool, error) {
	if _, err := os.Stat(p.ProjectName); err == nil {
		var overwrite bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Folder %s already exists. Overwrite?", p.ProjectName),
			Help:    "This will remove the existing folder and create a new project.",
		}
		if err := survey.AskOne(prompt, &overwrite); err != nil || !overwrite {
			return false, err
		}
		os.RemoveAll(p.ProjectName)
	}

	repo := config.RepoBase
	if repoURL != "" {
		repo = repoURL
	}

	fmt.Printf("Cloning repository: %s\n", repo)
	cmd := exec.Command("git", "clone", repo, p.ProjectName)
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		return false, err
	}
	return true, nil
}

func (p *Project) replacePackageName() error {
	packageName := helper.GetProjectName(p.ProjectName)
	return p.replaceFiles(packageName)
}

func (p *Project) configureProject() error {
	configurations := map[string]string{
		"{{FRAMEWORK}}": p.Framework,
		"{{DATABASE}}":  p.Database,
		"{{ORM}}":       p.ORM,
	}
	for placeholder, value := range configurations {
		if err := p.replaceFilesWithPlaceholder(placeholder, value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) replaceFiles(placeholder string) error {
	return filepath.Walk(p.ProjectName, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".go" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		newData := bytes.ReplaceAll(data, []byte(placeholder), []byte(p.ProjectName))
		return os.WriteFile(path, newData, 0644)
	})
}

func (p *Project) replaceFilesWithPlaceholder(placeholder, value string) error {
	return filepath.Walk(p.ProjectName, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		newData := bytes.ReplaceAll(data, []byte(placeholder), []byte(value))
		return os.WriteFile(path, newData, 0644)
	})
}

func (p *Project) modTidy() error {
	fmt.Println("Running go mod tidy")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = p.ProjectName
	return cmd.Run()
}

func (p *Project) rmGit() {
	os.RemoveAll(filepath.Join(p.ProjectName, ".git"))
}
