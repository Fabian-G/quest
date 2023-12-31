package di

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

var InternalEditTag = "quest-object-id"

type StyleDef struct {
	If string `mapstructure:"if,omitempty"`
	Fg string `mapstructure:"fg,omitempty"`
}

type TagDef struct {
	Type     string     `mapstructure:"type,omitempty"`
	Humanize bool       `mapstructure:"humanize,omitempty"`
	Styles   []StyleDef `mapstructure:"styles,omitempty"`
}

type MacroDef struct {
	Name       string   `mapstructure:"name,omitempty"`
	Query      string   `mapstructure:"query,omitempty"`
	InTypes    []string `mapstructure:"args,omitempty"`
	ResultType string   `mapstructure:"result,omitempty"`
	InjectIt   bool     `mapstructure:"inject-it,omitempty"`
}

func (m MacroDef) InDTypes() []qselect.DType {
	dtypes := make([]qselect.DType, 0, len(m.InTypes))
	for _, t := range m.InTypes {
		dtypes = append(dtypes, qselect.DType(t))
	}
	return dtypes
}

type ViewDef struct {
	Query       string   `mapstructure:"query,omitempty"`
	Projection  []string `mapstructure:"projection,omitempty"`
	Sort        []string `mapstructure:"sort,omitempty"`
	Clean       []string `mapstructure:"clean,omitempty"`
	Interactive bool     `mapstructure:"interactive,omitempty"`
	AddPrefix   string   `mapstructure:"add-prefix,omitempty"`
	AddSuffix   string   `mapstructure:"add-suffix,omitempty"`
}

type Config struct {
	TodoFile    string   `mapstructure:"todo-file,omitempty"`
	DoneFile    string   `mapstructure:"done-file,omitempty"`
	KeepBackups int      `mapstructure:"backup"`
	Editor      string   `mapstructure:"editor,omitempty"`
	UnknownTags bool     `mapstructure:"unknown-tags,omitempty"`
	ClearOnDone []string `mapstructure:"clear-on-done,omitempty"`
	QuestScore  struct {
		MinPriority  string `mapstructure:"min-priority,omitempty"`
		UrgencyTag   string `mapstructure:"urgency-tag,omitempty"`
		UrgencyBegin int    `mapstructure:"urgency-begin,omitempty"`
	} `mapstructure:"quest-score,omitempty"`
	Recurrence struct {
		RecTag       string `mapstructure:"rec-tag,omitempty"`
		DueTag       string `mapstructure:"due-tag,omitempty"`
		ThresholdTag string `mapstructure:"threshold-tag,omitempty"`
	} `mapstructure:"recurrence,omitempty"`
	Styles      []StyleDef         `mapstructure:"styles"`
	DefaultView ViewDef            `mapstructure:"default-view,omitempty"`
	Views       map[string]ViewDef `mapstructure:"views,omitempty"`
	Macros      []MacroDef         `mapstructure:"macro,omitempty"`
	Tags        map[string]TagDef  `mapstructure:"tags,omitempty"`
	NowFunc     func() time.Time   `mapstructure:"now-func,omitempty"` // Manually set only in testing, but defaults to time.Now
}

func (c Config) HumanizedTags() []string {
	humanizedTags := make([]string, 0)
	for key, tagDef := range c.Tags {
		if tagDef.Humanize {
			humanizedTags = append(humanizedTags, key)
		}
	}
	return humanizedTags
}

func (c Config) TagTypes() map[string]qselect.DType {
	tagTypes := make(map[string]qselect.DType)
	for key, tagDef := range c.Tags {
		tagTypes[key] = qselect.DType(tagDef.Type)
	}
	return tagTypes
}

func (c Config) TagColors() map[string]qprojection.ColorFunc {
	tagColors := make(map[string]qprojection.ColorFunc)
	for t, tagDef := range c.Tags {
		tagDef := tagDef
		ifs := make([]qselect.Func, 0)
		for i, s := range tagDef.Styles {
			f, err := qselect.CompileQQL(s.If)
			if err != nil {
				log.Fatal(fmt.Errorf("could not compile tag-color condition for %s #%d: %w", t, i, err))
			}
			ifs = append(ifs, f)
		}
		tagColors[t] = func(list *todotxt.List, item *todotxt.Item) *lipgloss.Color {
			for i, f := range ifs {
				if f(list, item) {
					c := lipgloss.Color(tagDef.Styles[i].Fg)
					return &c
				}
			}
			return nil
		}
	}
	return tagColors
}

func (c Config) LineColors() qprojection.ColorFunc {
	ifs := make([]qselect.Func, 0)
	for i, s := range c.Styles {
		f, err := qselect.CompileQQL(s.If)
		if err != nil {
			log.Fatal(fmt.Errorf("could not compile line color condition #%d: %w", i, err))
		}
		ifs = append(ifs, f)
	}
	return func(list *todotxt.List, item *todotxt.Item) *lipgloss.Color {
		for i, f := range ifs {
			if f(list, item) {
				c := lipgloss.Color(c.Styles[i].Fg)
				return &c
			}
		}
		return nil
	}
}

func buildConfig(file string) (Config, error) {
	v := viper.New()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(fmt.Errorf("could not determine users home directory: %w", err))
	}

	configHome := getConfigHome(homeDir)
	dataHome := getDataHome(homeDir)

	v.SetConfigType("toml")
	if file != "" {
		v.SetConfigFile(file)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(path.Join(configHome, "quest"))
	}

	err = v.ReadInConfig()
	var notFound viper.ConfigFileNotFoundError
	if err != nil && !errors.As(err, &notFound) {
		return Config{}, err
	}
	setDefaults(v, homeDir, dataHome)

	config := Config{}
	if err := v.UnmarshalExact(&config); err != nil {
		return Config{}, err
	}
	config.TodoFile = os.ExpandEnv(config.TodoFile)
	config.DoneFile = os.ExpandEnv(config.DoneFile)
	config.Tags[InternalEditTag] = TagDef{
		Type:     "int",
		Humanize: false,
	}

	return config, nil
}

func getConfigHome(home string) string {
	switch runtime.GOOS {
	case "linux":
		if cHome, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
			return cHome
		}
		return path.Join(home, ".config")
	case "windows":
		if cHome, ok := os.LookupEnv("APPDATA"); ok {
			return cHome
		}
		return path.Join(home, "AppData", "Roaming")
	case "darwin":
		return path.Join(home, "Library", "Application Support")
	}
	log.Fatalln("os not supported")
	return ""
}

func getDataHome(home string) string {
	switch runtime.GOOS {
	case "linux":
		if dHome, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
			return dHome
		}
		return path.Join(home, ".local", "share")
	case "windows":
		if dHome, ok := os.LookupEnv("LOCALAPPDATA"); ok {
			return dHome
		}
		return path.Join(home, "AppData", "Local")
	case "darwin":
		return path.Join(home, "Library")
	}
	log.Fatalln("os not supported")
	return ""
}

func setDefaults(v *viper.Viper, homeDir string, dataHome string) {
	v.SetDefault("todo-file", path.Join(dataHome, "quest/todo.txt"))
	v.SetDefault("done-file", path.Join(dataHome, "quest/done.txt"))
	v.SetDefault("backup", 0)
	v.SetDefault("editor", getDefaultEditor())
	v.SetDefault("unknown-tags", true)
	v.SetDefault("quest-score.urgency-tag", "due")
	v.SetDefault("quest-score.urgency-begin", 90)
	v.SetDefault("quest-score.min-priority", "E")
	v.SetDefault("clear-on-done", nil)
	v.SetDefault("recurrence.due-tag", "due")
	v.SetDefault("recurrence.threshold-tag", "t")
	v.SetDefault("default-view.query", "")
	v.SetDefault("default-view.projection", qprojection.StarProjection)
	v.SetDefault("default-view.sort", nil)
	v.SetDefault("default-view.clean", nil)
	v.SetDefault("default-view.interactive", true)
	v.SetDefault("default-view.add-prefix", "")
	v.SetDefault("default-view.add-suffix", "")
	v.SetDefault("tags", make(map[string]TagDef))
	v.SetDefault("now-func", time.Now)

	for viewName := range v.GetStringMap("views") {
		v.SetDefault("views."+viewName+".query", v.GetString("default-view.query"))
		v.SetDefault("views."+viewName+".projection", v.GetStringSlice("default-view.projection"))
		v.SetDefault("views."+viewName+".sort", v.GetStringSlice("default-view.sort"))
		v.SetDefault("views."+viewName+".clean", v.GetStringSlice("default-view.clean"))
		v.SetDefault("views."+viewName+".interactive", v.GetBool("default-view.interactive"))
		v.SetDefault("views."+viewName+".add-prefix", v.GetString("default-view.add-prefix"))
		v.SetDefault("views."+viewName+".add-suffix", v.GetString("default-view.add-suffix"))
	}
}

func getDefaultEditor() string {
	var possibleEditors []string = []string{os.Getenv("EDITOR"), "nano", "nvim", "vim", "vi", "emacs", "notepad.exe"}
	for _, editor := range possibleEditors {
		if p, err := exec.LookPath(editor); err == nil {
			return p
		}
	}
	return ""
}
