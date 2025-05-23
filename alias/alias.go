package alias

import (
	"github.com/chainreactors/fingers/common"
	"github.com/chainreactors/fingers/resources"
	"gopkg.in/yaml.v3"
	"strings"
)

func NewAliases(origin ...*Alias) (*Aliases, error) {
	var aliases []*Alias
	err := yaml.Unmarshal(resources.AliasesData, &aliases)
	if err != nil {
		return nil, err
	}
	aliasMap := &Aliases{
		Aliases: make(map[string]*Alias, len(aliases)+len(origin)),
		Map:     make(map[string]map[string]string),
	}

	err = aliasMap.Compile(append(origin, aliases...)) // yaml的优先级高于origin
	if err != nil {
		return nil, err
	}
	return aliasMap, nil
}

type Aliases struct {
	Aliases map[string]*Alias
	Map     map[string]map[string]string // 加速查询的映射表
}

func (as *Aliases) AppendAliases(other []*Alias) {
	err := as.Compile(other)
	if err != nil {
		return
	}
}

func (as *Aliases) Compile(aliases []*Alias) error {
	for _, alias := range aliases {
		alias.Name = strings.ToLower(alias.Name)
		alias.normalizedName = resources.NormalizeString(alias.Name)
		alias.blocked = make(map[string]bool)
		as.Aliases[alias.Name] = alias
		for _, block := range alias.Block {
			alias.blocked[block] = true
		}

		// 生成映射表
		for engine, engineMap := range alias.AliasMap {
			if _, ok := as.Map[engine]; !ok {
				as.Map[engine] = make(map[string]string)
			}
			for _, name := range engineMap {
				as.Map[engine][resources.NormalizeString(name)] = alias.Name
			}
		}
	}

	return nil
}

func (as *Aliases) Find(engine, name string) (*Alias, bool) {
	if engineMap, ok := as.Map[engine]; ok {
		if aliasName, ok := engineMap[name]; ok {
			if alias, ok := as.Aliases[aliasName]; ok {
				if !alias.blocked[engine] {
					return alias, true
				}
				return alias, false
			}
		}
	}
	return nil, false
}

func (as *Aliases) FindAny(name string) (string, *Alias, bool) {
	name = resources.NormalizeString(name)
	for engine, _ := range as.Map {
		alias, ok := as.Find(engine, name)
		if ok {
			return engine, alias, ok
		}
	}
	return "", nil, false
}

func (as *Aliases) FindFramework(frame *common.Framework) (*Alias, bool) {
	return as.Find(frame.From.String(), resources.NormalizeString(frame.Name))
}

type Alias struct {
	Name           string `json:"name" yaml:"name"`
	normalizedName string
	Vendor         string              `json:"vendor" yaml:"vendor"`
	Product        string              `json:"product" yaml:"product"`
	Version        string              `json:"version,omitempty" yaml:"version"`
	Update         string              `json:"update,omitempty" yaml:"update"`
	Edition        string              `json:"edition,omitempty" yaml:"edition"`
	AliasMap       map[string][]string `json:"alias" yaml:"alias"`
	Block          []string            `json:"block,omitempty" yaml:"block"`
	blocked        map[string]bool
}

func (a *Alias) IsBlocked(key string) bool {
	return a.blocked[key]
}

func (a *Alias) FuzzyMatch(s string) bool {
	return a.normalizedName == resources.NormalizeString(s)
}

func (a *Alias) ToWFN() *common.Attributes {
	return &common.Attributes{
		Part:    "a",
		Vendor:  a.Vendor,
		Product: a.Product,
		Version: a.Version,
		Update:  a.Update,
		Edition: a.Edition,
	}
}
