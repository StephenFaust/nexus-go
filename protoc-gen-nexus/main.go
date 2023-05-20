package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func main() {
	var flags flag.FlagSet
	genSer := flags.Bool("gen-ser", false, "--nexus_out=gen-ser=true:.")
	genCli := flags.Bool("gen-cli", false, "--nexus_out=gen-cli=true:.")
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(p *protogen.Plugin) error {
		var err error
		if *genCli {
			err = generateClient(p)
		}
		if *genSer {
			err = generateServer(p)
		}
		return err
	})
}

// generateServer generate service code
func generateServer(plugin *protogen.Plugin) error {
	for _, f := range plugin.Files {
		if len(f.Services) == 0 {
			continue
		}
		fileName := f.GeneratedFilenamePrefix + ".server.go"
		t := plugin.NewGeneratedFile(fileName, f.GoImportPath)
		t.P("// Code generated by protoc-gen-nexus.")
		t.P()
		pkg := fmt.Sprintf("package %s", f.GoPackageName)
		t.P(pkg)
		t.P()
		importCode := `import (
		"github.com/StephenFaust/nexus-go/nexus"
        "reflect"
		)`
		t.P(importCode)
		t.P()
		for _, s := range f.Services {
			serviceCode := fmt.Sprintf(`%stype %s struct{}`,
				getComments(s.Comments, string(s.Desc.Name())), s.Desc.Name())
			t.P(serviceCode)
			t.P()
			for _, m := range s.Methods {
				funcCode := fmt.Sprintf(`%sfunc(s *%s) %s(args *%s,reply *%s)error{
					// define your service ...
					return nil
				}
				`, getComments(m.Comments, string(m.Desc.Name())), s.Desc.Name(),
					m.Desc.Name(), m.Input.Desc.Name(), m.Output.Desc.Name())
				t.P(funcCode)
			}
			t.P(fmt.Sprintf(`//Get%sInfo method generated by protoc-gen-nexus.`, s.Desc.Name()))
			serviceInfoCode := fmt.Sprintf(`func Get%sInfo() nexus.ServiceInfo{
                   service := &%s{}
	               serviceType := reflect.TypeOf(service)
	               methodMap := make(map[string]*nexus.MethodInfo)
	               for i := 0; i < serviceType.NumMethod(); i++ {
		               method := serviceType.Method(i)
		               methodMap[method.Name] = &nexus.MethodInfo{
			                   Method:   method,
			                   ParType:  method.Type.In(1),
			                   RelyType: method.Type.In(2),
		               }
	          }
	             return nexus.ServiceInfo{
                      Name:    "%s",
		              Value:   reflect.ValueOf(service),
		              Methods: methodMap,
	              }
    }
`, s.Desc.Name(), s.Desc.Name(), s.Desc.Name())
			t.P(serviceInfoCode)
			t.P()

		}
	}
	return nil
}

func generateClient(plugin *protogen.Plugin) error {
	for _, f := range plugin.Files {
		if len(f.Services) == 0 {
			continue
		}
		fileName := f.GeneratedFilenamePrefix + ".client.go"
		t := plugin.NewGeneratedFile(fileName, f.GoImportPath)
		t.P("// Code generated by protoc-gen-nexus.")
		t.P()
		pkg := fmt.Sprintf("package %s", f.GoPackageName)
		t.P(pkg)
		t.P()
		importCode := `import (
		"github.com/StephenFaust/nexus-go/nexus"
		)`
		t.P(importCode)
		t.P()
		for _, s := range f.Services {
			clientName := fmt.Sprintf(`%sClient`, s.Desc.Name())
			clientCode := fmt.Sprintf(`%stype %s struct{
                 client *nexus.RpcClient
			}
			`,
				getComments(s.Comments, clientName), clientName)
			t.P(clientCode)
			t.P()
			t.P(fmt.Sprintf(`func New%sClient(client *nexus.RpcClient) %sClient {
	                return %sClient{
		                client: client,
	             }
            }`, s.Desc.Name(), s.Desc.Name(), s.Desc.Name()))
			for _, m := range s.Methods {
				funcCode := fmt.Sprintf(`%sfunc(sc *%sClient) %s(args *%s,reply *%s)error{
                    return sc.client.Invoke(args, "%s", "%s", reply)
				}
				`, getComments(m.Comments, string(m.Desc.Name())), s.Desc.Name(),
					m.Desc.Name(), m.Input.Desc.Name(), m.Output.Desc.Name(), s.Desc.Name(), m.Desc.Name())
				t.P(funcCode)
			}
		}
	}
	return nil
}

// cd getComments get comment details
func getComments(comments protogen.CommentSet, name string) string {
	c := make([]string, 0)
	c = append(c, strings.Split(string(comments.Leading), "\n")...)
	c = append(c, strings.Split(string(comments.Trailing), "\n")...)

	res := ""
	for _, comment := range c {
		if strings.TrimSpace(comment) == "" {
			continue
		}
		res += "//" + name + " " + comment + "\n"
	}
	return res
}
