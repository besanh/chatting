package model

import "fmt"

type Param struct {
	Key      string
	Value    string
	Operator string
}

func (p *Param) BuildQuery() string {
	if len(p.Key) < 1 || len(p.Value) < 1 || p.Value == "%%" {
		return ""
	} else if len(p.Operator) < 1 {
		p.Operator = "="
	}
	return fmt.Sprintf("%s %s ?", p.Key, p.Operator)
}
