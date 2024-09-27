package markdown_parser

import (
	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

func ParseMetadata(yamlData string) (*m.ExplanationMetaData, error)  {
    var res *m.ExplanationMetaData
    err := yaml.Unmarshal([]byte(yamlData), &res)
    if err != nil {
        return nil, err
    }

	return res, err
}

func ParseBody(markdown string) ([]byte) {
    return blackfriday.Run([]byte(markdown))
}