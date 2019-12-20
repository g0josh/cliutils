package main

import (
    "log"
    "os"
    "os/exec"
    "io/ioutil"
    "strings"
    "strconv"
    "reflect"
    "gopkg.in/yaml.v3"
)
const (
    NET_DIR string = "/sys/class/net"
    THEME_PATH string = "~/.config/themes/theme"
    POLY_INFO_PATH string = "/tmp/polybar_info1"
)
var POWER_ICONS = map[string]string{
        "power":"","reboot":"","lock":"",
        "logout":"", "cancel":"",
    }
type Theme struct {
    TerminalColors map[string]string `yaml:"terminal_colors"`
    BodyFg string `yaml:"bodyfg"`
    BodyBg string `yaml:"bodybg"`
    TitleBg string `yaml:"titlebg"`
    TitleFg string `yaml:"titlefg"`
    OccupiedFg string `yaml:"occupiedfg"`
    OccupiedBg string `yaml:"occupiedbg"`
    FocusedFg string `yaml:"focusedfg"`
    FocusedBg string `yaml:"focusedbg"`
    UrgentFg string `yaml:"urgentfg"`
    UrgentBg string `yaml:"urgentbg"`
    AltFg string `yaml:"altfg"`
    AltBg string `yaml:"altbg"`
    ModuleSeparator string `yaml:"moduleseparator"`
    WsPadding int `yaml:"wspadding"`
    Background string `yaml:"background"`
    LeftModulePrefix string `yaml:"leftmoduleprefix"`
    LeftModuleSuffix string `yaml:"leftmodulesuffix"`
    RightModulePrefix string `yaml:"rightmoduleprefix"`
    RightModuleSuffix string `yaml:"rightmodulesuffix"`
    TitlePadding int `yaml:"titlepadding"`
    BodyPadding int `yaml:"bodypadding"`
    FocusedWindowBorder string `yaml:"focusedwindowborder"`
}

func getInterfaces() (string, string, string) {
    files, err := ioutil.ReadDir(NET_DIR)
    if err != nil {
        log.Fatal(err)
    }
    lan1, lan2, wlan := "", "", ""
    for _, file := range files {
        if strings.HasPrefix(file.Name(), "e") {
            if lan1 != "" {
                lan2 = file.Name()
            } else {
                lan1 = file.Name()
            }
        } else if strings.HasPrefix(file.Name(), "w") {
            wlan = file.Name()
        }
    }
    return lan1, lan2, wlan
}

func setupScreens() ([]string) {
    _cmd := exec.Command("xrandr")
    stdoutStderr, err := _cmd.CombinedOutput()
    if err != nil {
        log.Fatal("Error while xrandr\n%v", err)
    }
    lines := strings.Split(string(stdoutStderr), "\n")
    var(
        connected []string
        cmd []string
        x = 0
    )
    for i, line := range lines {
        if !strings.Contains(line, "connected") {
            continue
        }
        name := strings.Fields(line)[0]
        if strings.Contains(line, " connected") {
            res := strings.Fields(lines[i+1])[0]
            cmd = append(cmd, "--output", name, "--mode", res,"--pos", strconv.Itoa(x) + "x0", "--rotate", "normal")
            _x, err := strconv.Atoi(strings.Split(res, "x")[0])
            if err != nil {
                log.Fatalf("Error while Atoi\n%v", err)
            }
            x = x + _x
            connected = append(connected, name)
        } else if strings.Contains(line, "disconnected") {
            cmd = append(cmd, "--output", name, "--off")
        }
    }
    _cmd = exec.Command("xrandr", cmd...)
    err = _cmd.Start()
    if err != nil {
        log.Fatalf("Error while running %v\n%v", cmd, err)
    }
    return connected
}

func main() {
    _homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("Error while getting home directory\n%v",err)
    }
    theme_path := strings.Replace(THEME_PATH, "~", _homeDir, 1)
    yamlFh, err := ioutil.ReadFile(theme_path)
    if err != nil {
        log.Fatalf("Error while opening theme file '%v'\n%v",theme_path, err)
    }
    var theme Theme
    err = yaml.Unmarshal(yamlFh, &theme)
    if err != nil {
        log.Fatalf("Error while Unmarshaling yaml file '%v'\n%v",theme_path, err)
    }
    if theme.OccupiedBg == "" {
        theme.OccupiedBg = theme.BodyBg
    }
    if theme.OccupiedFg == "" {
        theme.OccupiedFg = theme.BodyFg
    }
    // Polybar Ws formats
    formats := make(map[string]string)
    formats["layoutWs"] = "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.LeftModuleSuffix+"%{F-}%{B-}"
    formats["activeWs"] = "%{B"+theme.Background+"}%{F"+theme.FocusedBg+"}"+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.FocusedBg+"}%{F"+theme.FocusedFg+"}"+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.FocusedBg+"}"+theme.LeftModuleSuffix+"%{F-}%{B-}"
    formats["activeWsOther"] = "%{B"+theme.Background+"}%{F"+theme.BodyBg+"}"+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.BodyBg+"}%{F"+theme.UrgentBg+"}"+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.BodyBg+"}"+theme.LeftModuleSuffix+"%{F-}%{B-}"
    formats["occupiedWs"] = "%{B"+theme.Background+"}%{F"+theme.OccupiedBg+"}"+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.OccupiedBg+"}%{F"+theme.OccupiedFg+"}"+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.OccupiedBg+"}"+theme.LeftModuleSuffix+"%{F-}%{B-}"
    formats["visibleWs"] = "%{B"+theme.Background+"}%{F"+theme.AltBg+"}"+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.AltBg+"}%{F"+theme.AltFg+"}"+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.AltBg+"}"+theme.LeftModuleSuffix+"%{F-}%{B-}"
    formats["visibleWsOther"] = "%{B"+theme.Background+"}%{F"+theme.BodyBg+"}"+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.BodyBg+"}%{F"+theme.AltBg+"}"+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.BodyBg+"}"+theme.LeftModuleSuffix+"%{F-}%{B-}"
    formats["urgetWs"] = "%{B"+theme.Background+"}%{F"+theme.UrgentBg+""+theme.LeftModulePrefix+"%{F-}%{B-}%{B"+theme.UrgentBg+"%{F"+theme.UrgentFg+""+strings.Repeat(" ",theme.WsPadding)+"%label%"+strings.Repeat(" ",theme.WsPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.UrgentBg+""+theme.LeftModuleSuffix+"%{F-}%{B-}"

    // Other Poly vars
    poly_vars := make(map[string]string)
    poly_vars["poweropen"]= "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.TitlePadding)+POWER_ICONS["power"]+strings.Repeat(" ",theme.TitlePadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModuleSuffix+"%{F-}%{B-}"
    poly_vars["powerclose"]= "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.TitlePadding)+POWER_ICONS["cancel"]+strings.Repeat(" ",theme.TitlePadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModuleSuffix+"%{F-}%{B-}"
    poly_vars["reboot"]= "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.BodyPadding)+POWER_ICONS["reboot"]+strings.Repeat(" ",theme.BodyPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModuleSuffix+"%{F-}%{B-}"
    poly_vars["powerof"]= "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.BodyPadding)+POWER_ICONS["power"]+strings.Repeat(" ",theme.BodyPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModuleSuffix+"%{F-}%{B-}"
    poly_vars["logout"]= "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.BodyPadding)+POWER_ICONS["logout"]+strings.Repeat(" ",theme.BodyPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModuleSuffix+"%{F-}%{B-}"
    poly_vars["lock"]= "%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModulePrefix+"%{F-}%{B-}%{B"+theme.TitleBg+"}%{F"+theme.TitleFg+"}"+strings.Repeat(" ",theme.BodyPadding)+POWER_ICONS["lock"]+strings.Repeat(" ",theme.BodyPadding)+"%{F-}%{B-}%{B"+theme.Background+"}%{F"+theme.TitleBg+"}"+theme.RightModuleSuffix+"%{F-}%{B-}"
    //log.Printf("formats\n%+v\n%+v", formats, poly_vars)
    lan1, lan2, wlan := getInterfaces()
    connectedScreens := setupScreens()
    //log.Printf("Interfaces:%s, %s, %s\nConnected screens:%v\ntheme\n%+v",lan1, lan2, wlan, connectedScreens, theme)
    type PolybarInfo struct{
        ConnectedScreens map[int]map[string]string
        PolybarWsFormats map[string]string
        PolybarModuleSeparator string
    }
    var polybarInfo PolybarInfo
    connectedScreenMap := make(map[int]map[string]string)
    themeValues := reflect.ValueOf(theme)
    themeType := themeValues.Type()
    err = exec.Command("killall", "polybar").Run()
    for i, screen := range connectedScreens {
        os.Setenv("POLY_MONITOR", screen)
        os.Setenv("POLY_POWER_OPEN", poly_vars["poweropen"])
        os.Setenv("POLY_POWER_CLOSE", poly_vars["powerclose"])
        os.Setenv("POLY_POWEROFF", poly_vars["poweroff"])
        os.Setenv("POLY_REBOOT", poly_vars["reboot"])
        os.Setenv("POLY_LOGOUT", poly_vars["logout"])
        os.Setenv("POLY_LOCK", poly_vars["lock"])
        os.Setenv("POLY_WLAN", wlan)
        os.Setenv("POLY_LAN1", lan1)
        os.Setenv("POLY_LAN2", lan2)
        for i:=0; i<themeValues.NumField(); i++ {
            _key := "POLY_" + strings.ToUpper(themeType.Field(i).Name)
            _value := themeValues.Field(i).Interface()
            if reflect.TypeOf(_value).Kind() == reflect.Int {
                value := strconv.Itoa(_value.(int))
                os.Setenv(_key, value)
            } else if reflect.TypeOf(_value).Kind() == reflect.String {
                os.Setenv(_key, _value.(string))
            }
        }
        _cmd := exec.Command("polybar","-r","island")
        err = _cmd.Start()
        if err != nil {
            log.Fatalf("error while launching polybar\n%v",err)
        }
        connectedScreenMap[i] = map[string]string{"name":screen,
            "pid":strconv.Itoa(_cmd.Process.Pid)}
    }
    polybarInfo.ConnectedScreens = connectedScreenMap
    polybarInfo.PolybarWsFormats = formats
    polybarInfo.PolybarModuleSeparator = theme.ModuleSeparator
    yamlFile, err := yaml.Marshal(&polybarInfo)
    if err != nil {
        log.Fatalf("error while marshalling:\n%v", err)
    }
    err = ioutil.WriteFile(POLY_INFO_PATH, yamlFile, 0644)
    if err != nil {
        log.Fatalf("Error while writing polybar info file:'%v'\n%v",POLY_INFO_PATH, err)
    }
}
