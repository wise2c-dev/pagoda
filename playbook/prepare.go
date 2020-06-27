package playbook

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
)

// Template src is template path and dest is template output file
type Template struct {
	Src  string `json:"-"`
	Dest string `json:"-"`
}

const (
	//PlaybookSuffix - suffix for playbook folder
	PlaybookSuffix  = "-playbook"
	ansibleGroupDir = "group_vars"
	hostsTmpl       = "hosts.gotmpl"
	hostsFile       = "hosts"
	tmplDir         = "yat"
	tmplSuffix      = ".gotmpl"
)

func PreparePlaybooks(dir string, ds *DeploySeed) error {
	for k, v := range map[string]*Component(*ds) {
		clusterID := fmt.Sprintf("%v", v.Inherent["cluster_id"])
		playbookRootPath := path.Join(dir, k+PlaybookSuffix, v.Version)
		if err := preparePlaybook(clusterID, playbookRootPath, ds); err != nil {
			return err
		}
	}

	return nil
}

func preparePlaybook(clusterID, pbRootPath string, ds *DeploySeed) error {
	tps, err := getTemplatePath(clusterID, pbRootPath)
	if err != nil {
		return err
	}

	for _, tp := range tps {
		if err = applyTemplate(tp, ds); err != nil {
			return err
		}
	}

	return nil
}

func applyTemplate(t *Template, ds *DeploySeed) error {
	file, err := os.OpenFile(t.Dest, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create template dest file %s error: %s", t.Dest, err)
	}
	defer file.Close()

	content, err := ioutil.ReadFile(t.Src)
	if err != nil {
		return fmt.Errorf("read template src file error: %v", err)
	}
	if strings.Contains(t.Src, tmplSuffix) {
		tp := template.Must(template.New("ansible").Funcs(fns).Parse(string(content)))
		// tp.Option("missingkey=zero")
		err = tp.Execute(file, ds)
		if err != nil {
			return fmt.Errorf("execute template for %s error: %v", t.Dest, err)
		}
	} else {
		if _, err = file.Write(content); err != nil {
			return fmt.Errorf("write file for %s error: %v", t.Dest, err)
		}
	}
	return nil
}

// getTemplatePath - check playbook, get every template path & output file
func getTemplatePath(clusterID, bpRootPath string) ([]*Template, error) {
	if err := createClusterPath(clusterID, bpRootPath); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(path.Join(bpRootPath, tmplDir))
	if err != nil {
		return nil, fmt.Errorf("read dir %s error: %v", bpRootPath, err)
	}

	tps := make([]*Template, 0, len(files))
	for _, f := range files {
		var dest string
		if !strings.Contains(f.Name(), tmplSuffix) {
			dest = path.Join(bpRootPath, "clusters", clusterID, f.Name())
		}
		if f.Name() == hostsTmpl {
			dest = path.Join(bpRootPath, "clusters", clusterID, hostsFile)
		} else {
			dest = path.Join(bpRootPath, "clusters", clusterID, ansibleGroupDir, strings.TrimSuffix(f.Name(), tmplSuffix))
		}
		t := &Template{
			Src:  path.Join(bpRootPath, tmplDir, f.Name()),
			Dest: dest,
		}
		tps = append(tps, t)
	}

	return tps, nil
}

func createClusterPath(clusterID, pbRootPath string) error {
	// playbookRootPath: /{componentName}-playbook/{version}
	// 01. check template path: /{playbookRootPath}/yat
	tmplFiles, err := ioutil.ReadDir(path.Join(pbRootPath, tmplDir))
	if err != nil {
		return fmt.Errorf("check path %s error: %v", pbRootPath, err)
	}
	// 02. check path: /{playbookRootPath}/clusters/{clusterID}, if not exist, create it.
	clusterPath := fmt.Sprintf("%s/clusters/%s", pbRootPath, clusterID)
	if _, err := os.Stat(clusterPath); os.IsNotExist(err) {
		if err = os.MkdirAll(clusterPath, 0755); err != nil {
			return fmt.Errorf("create cluster path %s error: %v", clusterPath, err)
		}
	}
	// 03. if check group vars is exist, create path: /{playbookRootPath}/clusters/{clusterID}/group_vars/
	var hasGroupVarsPath, hasHostsTmpl bool
	for _, f := range tmplFiles {
		if f.Name() != hostsTmpl && !hasGroupVarsPath {
			if err = os.MkdirAll(clusterPath + "/" + ansibleGroupDir, 0755); err != nil {
				return fmt.Errorf("create group vars path %s error: %v", clusterPath, err)
			}
			hasGroupVarsPath = true
		} else if f.Name() == hostsTmpl {
			hasHostsTmpl = true
		}
	}
	// 04. check hosts template
	if !hasHostsTmpl {
		return fmt.Errorf("not found %s", hostsTmpl)
	}
	return nil
}

func GetFileFromDir(dir string, cf func(os.FileInfo) bool) (fs []os.FileInfo, err error) {
	files, err := ioutil.ReadDir(dir)

	for _, f := range files {
		if cf(f) {
			fs = append(fs, f)
		}
	}

	return
}

var fns = template.FuncMap{
	"notLast": func(x int, a interface{}) bool {
		return x < reflect.ValueOf(a).Len()-1
	},
}
