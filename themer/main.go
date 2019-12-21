package main

import (
    "log"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "io/ioutil"
    "gopkg.in/yaml.v3"
)

const (
    AlacrittyConfPath = "~/.config/alacritty/alacritty.yml"
    XColorsPath = "/tmp/x_colors"
    ThemePath = "~/.config/themes/current.theme"
    ParsedThemePath = "~/.config/themes/theme"
)
var ColorMap = map[string]string {"*.foreground:":"foreground","*.background:":"background",
        "*.cursorColor:":"cursorColor","*.color0:":"black", "*.color8:":"bright_black",
        "*.color1:":"red","*.color9:":"bright_red","*.color2:":"green",
        "*.color10:":"bright_green", "*.color3:":"yellow","*.color11:":"bright_yellow",
        "*.color4:":"blue", "*.color12:":"bright_blue", "*.color5:":"magenta",
        "*.color13:":"bright_magenta", "*.color6:":"cyan", "*.color14:":"bright_cyan",
        "*.color7:":"white","*.color15:":"bright_white"}
var homeDir string

func init() {
    var err error
    homeDir, err = os.UserHomeDir()
    if err != nil {
        log.Fatalf("Error while getting home directory\n%v",err)
    }
}

func getTheme() (map[string]interface{}) {
    themePath := strings.Replace(ThemePath, "~", homeDir, 1)
    yamlFh, err := ioutil.ReadFile(themePath)
    if err != nil {
        log.Fatalf("Error while opening theme file '%v'\n%v",themePath, err)
    }
    var theme = map[string]string{}
    err = yaml.Unmarshal(yamlFh, &theme)
    if err != nil {
        log.Fatalf("Error while Unmarshaling yaml file '%v'\n%v",themePath, err)
    }

    // get x colors and convert 
    // to human readable color keys
    colors := strings.Fields(theme["terminal_colors"])
    terminalColors := map[string]string{}
    xColors := ""
    for i:=0; i<len(colors); i+=2 {
        key := strings.TrimSpace(colors[i])
        if key == "!"{
            continue
        }
        color := strings.TrimSpace(colors[i+1])
        xColors += fmt.Sprintf("%s %s\n",key, color)
        if value, ok := ColorMap[key]; ok {
            terminalColors[value] = color
        }
    }
    xColors += fmt.Sprintf("rofi.color-window: #a0%s, %s, %s\n",terminalColors["red"][1:], terminalColors["background"], terminalColors["background"])
    xColors += fmt.Sprintf("rofi.color-normal: #00000000, %s, #00000000, %s, %s",terminalColors["background"], terminalColors["background"], terminalColors["red"])
    err = ioutil.WriteFile(XColorsPath, []byte(xColors), 0644)
    if err != nil {
        log.Fatalf("Error while writing file to %s:\n%v",XColorsPath, err)
    }

    //Generate and save theme file with
    //color codes everywhere
    correctedTheme := map[string]interface{} {
        "terminal_colors" : terminalColors,
    }
    for k, v := range theme {
        if k == "terminal_colors" {
            continue
        }
        if colorCode, ok := terminalColors[v]; ok{
            correctedTheme[k] = colorCode
        }else{
            correctedTheme[k] = v
        }
    }
    return correctedTheme
}

func main() {
    theme := getTheme()
    themeYaml, err := yaml.Marshal(&theme)
    if err != nil {
        log.Fatalf("Error while marshalling theme:\n%v", err)
    }
    //Save the parsed/corrected theme yaml
    themePath := strings.Replace(ParsedThemePath, "~", homeDir, 1)
    err = ioutil.WriteFile(themePath, themeYaml, 0644)
    if err != nil {
        log.Fatalf("Error while writing file to %s:\n%v",themePath, err)
    }

    //load x theme
    cmd := exec.Command("xrdb", "-merge", XColorsPath)
    if err := cmd.Start(); err != nil {
        log.Fatalf("Error while running command %v:\n%v",cmd, err)
    }

    //load alacritty yaml file
    yamlFh, err := ioutil.ReadFile(strings.Replace(AlacrittyConfPath, "~", homeDir, 1))
    if err != nil {
        log.Fatalf("error while reading %v\n:%v", AlacrittyConfPath, err)
    }
    alaConf := make(map[string]interface{})
    err = yaml.Unmarshal(yamlFh, &alaConf)
    if err != nil {
        log.Fatalf("Error while Unmarshaling yaml file '%v'\n%v",AlacrittyConfPath, err)
    }

    alaConfColors := struct {
        Primary map[string]string   `yaml:primary`
        Normal map[string]string    `yaml:normal`
        Bright map[string]string    `yaml:bright`
        Indexed_colors []string     `yaml:indexed_colors`
        } {
            Primary: make(map[string]string),
            Normal: make(map[string]string),
            Bright: make(map[string]string),
           }
    for key, value := range theme["terminal_colors"].(map[string]string){
        if strings.Contains(key, "ground") {
            alaConfColors.Primary[key] = value
        }else if strings.Contains(key, "bright") {
            key = strings.Split(key, "_")[1]
            alaConfColors.Bright[key] = value
        }else{
            alaConfColors.Normal[key] = value
        }
    }

    alaConf["colors"] = alaConfColors

    alaConfYaml, err := yaml.Marshal(&alaConf)
    if err != nil {
        log.Fatalf("Error while marshalling alacritty conf:\n%v", err)
    }
    alaConfPath := strings.Replace(AlacrittyConfPath, "~", homeDir, 1)
    err = ioutil.WriteFile(alaConfPath, alaConfYaml, 0644)
    if err != nil {
        log.Fatalf("Error while writing file to %s:\n%v",alaConfPath, err)
    }
}
