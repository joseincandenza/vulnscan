package code

import (
	"fmt"
	"github.com/simplycubed/vulnscan/adapters"
	"github.com/simplycubed/vulnscan/adapters/output"
	"github.com/simplycubed/vulnscan/entities"
	"github.com/simplycubed/vulnscan/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Analysis(command utils.Command, entity *entities.CodeAnalysis, adapter adapters.AdapterMap) {
	var analysisName = entities.Code
	_ = adapter.Output.Logger(output.ParseInfo(analysisName, "starting"))
	if walkErr := filepath.Walk(command.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".m" || filepath.Ext(path) == ".swift" {
			var jfilePath string
			// TODO: why are we doing this?
			if strings.Contains(filepath.Base(path), "+") {
				jfilePath = filepath.Join(filepath.Dir(path),
					strings.Replace(filepath.Base(path), "+", "x", -1))
				err := os.Rename(path, jfilePath)
				if err != nil {
					return fmt.Errorf("error moving file %s to %s: %s", path, jfilePath, err)
				}
			} else {
				jfilePath = path
			}

			var data string
			if d, err := ioutil.ReadFile(jfilePath); err != nil {
				return fmt.Errorf("error reading file %s: %s", jfilePath, err)
			} else {
				data = string(d)
			}
			relativeSrcPath := strings.Replace(jfilePath, command.Path, "", 1)
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "extracting rules..."))
			_ = ruleExtractor(data, relativeSrcPath, entity)
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "rules extracted!"))
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "extracting apis..."))
			_ = apiExtractor(data, relativeSrcPath, entity)
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "apis extracted!"))
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "extracting urls..."))
			_ = urlExtractor(data, relativeSrcPath, entity)
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "urls extracted!"))
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "extracting emails"))
			_ = emailExtractor(data, relativeSrcPath, entity)
			_ = adapter.Output.Logger(output.ParseInfo(analysisName, "emails extracted!"))
		}
		return nil
	}); adapter.Output.Error(output.ParseError(analysisName, walkErr)) != nil {
		return
	}
	if a := adapter.Services.MalwareDomains; a != nil {
		if adapter.Output.Error(output.ParseError(analysisName,a(command, entity))) != nil {
			return
		}
	}
	if err := adapter.Output.Result(command, entity); err != nil {
		_ = adapter.Output.Error(output.ParseError(analysisName, err))
	}
}


func ruleExtractor(data, path string, entity *entities.CodeAnalysis) entities.Entity {
	for _, rule := range Rules {
		if rule.Match(data) {
			var found bool
			for i, cf := range entity.Codes {
				if cf.Description == rule.Description {
					entity.Codes[i].Paths = append(entity.Codes[i].Paths, path)
					found = true
					break
				}
			}
			if !found {
				entity.Codes = append(entity.Codes, entities.CodeFinding{
					CodeRule: rule.CodeRule,
					Paths:    []string{path},
				})
			}
		}
	}
	return entity
}

func apiExtractor(data, path string, entity *entities.CodeAnalysis) entities.Entity {
	for _, api := range APIs {
		if api.Match(data) {
			var found bool
			for i, af := range entity.Apis {
				if af.Description == api.Description {
					entity.Apis[i].Paths = append(entity.Apis[i].Paths, path)
					found = true
					break
				}
			}
			if !found {
				entity.Apis = append(entity.Apis, entities.ApiFinding{
					ApiRule: api.ApiRule,
					Paths:    []string{path},
				})
			}
		}
	}
	return entity
}

func urlExtractor(data, path string, entity *entities.CodeAnalysis) entities.Entity {
	urlPat, _ := regexp.
		Compile(`https?://(?:[-\w.]|(?:%[\da-fA-F]{2}))+`)
	urls := urlPat.FindAll([]byte(data), -1)
	for _, url := range urls {
		var found bool
		for i, uf := range entity.Urls {
			if uf.Url == string(url) {
				entity.Urls[i].Paths = append(entity.Urls[i].Paths, path)
				found = true
				break
			}
		}
		if !found {
			entity.Urls = append(entity.Urls, entities.UrlFinding{
				Url: 	  string(url),
				Paths:    []string{path},
			})
		}
	}
	return entity
}

func emailExtractor(data, path string, entity *entities.CodeAnalysis) entities.Entity {
	emailPat, _ := regexp.Compile(`[\w.-]+@[\w-]+\.[\w.]+`)
	emails := emailPat.FindAll([]byte(data), -1)
	for email := range emails {
		var found bool
		for i, ef := range entity.Emails {
			if ef.Email == string(email) {
				entity.Emails[i].Paths = append(entity.Emails[i].Paths, path)
				found = true
				break
			}
		}
		if !found {
			entity.Emails = append(entity.Emails, entities.EmailFinding{
				Email: 	  string(email),
				Paths:    []string{path},
			})
		}
	}
	return entity
}

