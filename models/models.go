package models

type Explanation struct {
    Name string
    Title string
    Body []byte
}

type ExplanationMetaData struct {
    Title string `yaml:"title"`
    DateCreated string `yaml:"date_created"`
}