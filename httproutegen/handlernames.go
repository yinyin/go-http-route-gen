package httproutegen

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var defaultMethodEvaluateOrder = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodOptions}

// HandlerInvokeProfile represent invoke profile for given handler.
type HandlerInvokeProfile struct {
	RequestMethod string `json:"method"`
	HandlerName   string `json:"handler"`
	SameNext      bool   `json:"same-next"`
}

// HandlerNames represent handler names for one endpoint.
type HandlerNames struct {
	EvaluateOrder  []string                `yaml:"evaluate-order,omitempty" json:"evaluate-order,omitempty"`
	InvokeProfiles []*HandlerInvokeProfile `yaml:"-" json:"invoke-profile,omitempty"`
	GetHandler     string                  `yaml:"get,omitempty" json:"get,omitempty"`
	HeadHandler    string                  `yaml:"head,omitempty" json:"head,omitempty"`
	PostHandler    string                  `yaml:"post,omitempty" json:"post,omitempty"`
	PutHandler     string                  `yaml:"put,omitempty" json:"put,omitempty"`
	PatchHandler   string                  `yaml:"patch,omitempty" json:"patch,omitempty"`
	DeleteHandler  string                  `yaml:"delete,omitempty" json:"delete,omitempty"`
	OptionsHandler string                  `yaml:"options,omitempty" json:"options,omitempty"`
}

func (hn *HandlerNames) getHandlerNameByMethod(methodName string) string {
	methodName = strings.ToUpper(methodName)
	n := ""
	switch methodName {
	case http.MethodGet:
		n = hn.GetHandler
	case http.MethodHead:
		n = hn.HeadHandler
	case http.MethodPost:
		n = hn.PostHandler
	case http.MethodPut:
		n = hn.PutHandler
	case http.MethodPatch:
		n = hn.PatchHandler
	case http.MethodDelete:
		n = hn.DeleteHandler
	case http.MethodOptions:
		n = hn.OptionsHandler
	}
	return n
}

// cleanupEvaluateOrder also rebuild invoke profile
func (hn *HandlerNames) rebuildInvokeOrder() {
	evalOrder := hn.EvaluateOrder
	evalOrder = append(evalOrder, defaultMethodEvaluateOrder...)
	checkedStat := make(map[string]bool)
	var resultEvalOrder []string
	var invokeProfiles []*HandlerInvokeProfile
	for _, methodName := range evalOrder {
		methodName = strings.ToUpper(methodName)
		chk := checkedStat[methodName]
		if chk {
			continue
		}
		checkedStat[methodName] = true
		n := hn.getHandlerNameByMethod(methodName)
		if n != "" {
			resultEvalOrder = append(resultEvalOrder, methodName)
			aux := &HandlerInvokeProfile{
				RequestMethod: methodName,
				HandlerName:   n,
			}
			invokeProfiles = append(invokeProfiles, aux)
		}
	}
	for idx, profile := range invokeProfiles {
		if (idx + 1) >= len(invokeProfiles) {
			break
		}
		if profile.HandlerName == invokeProfiles[idx+1].HandlerName {
			profile.SameNext = true
		}
	}
	hn.EvaluateOrder = resultEvalOrder
	hn.InvokeProfiles = invokeProfiles
}

func (hn *HandlerNames) expandHandlerName(n string) (string, bool) {
	if "" == n {
		return n, false
	}
	if n[0] == '=' {
		switch strings.ToLower(n) {
		case "=get":
			n = hn.GetHandler
		case "=head":
			n = hn.HeadHandler
		case "=post":
			n = hn.PostHandler
		case "=put":
			n = hn.PutHandler
		case "=patch":
			n = hn.PatchHandler
		case "=delete":
			n = hn.DeleteHandler
		case "=options":
			n = hn.OptionsHandler
		default:
			log.Printf("WARN: unknown handler name assignment: %v", n)
			n = ""
		}
		return n, true
	}
	return n, false
}

func (hn *HandlerNames) expandNames() {
	runExpand := true
	remainExpandCycle := 6
	for runExpand && (remainExpandCycle > 0) {
		runExpand = false
		remainExpandCycle--
		var expanded bool
		hn.GetHandler, expanded = hn.expandHandlerName(hn.GetHandler)
		runExpand = runExpand || expanded
		hn.HeadHandler, expanded = hn.expandHandlerName(hn.HeadHandler)
		runExpand = runExpand || expanded
		hn.PostHandler, expanded = hn.expandHandlerName(hn.PostHandler)
		runExpand = runExpand || expanded
		hn.PutHandler, expanded = hn.expandHandlerName(hn.PutHandler)
		runExpand = runExpand || expanded
		hn.PatchHandler, expanded = hn.expandHandlerName(hn.PatchHandler)
		runExpand = runExpand || expanded
		hn.DeleteHandler, expanded = hn.expandHandlerName(hn.DeleteHandler)
		runExpand = runExpand || expanded
		hn.OptionsHandler, expanded = hn.expandHandlerName(hn.OptionsHandler)
		runExpand = runExpand || expanded
	}
}

func (hn *HandlerNames) cleanup() {
	if nil == hn {
		return
	}
	hn.expandNames()
	hn.rebuildInvokeOrder()
}

func (hn *HandlerNames) isEmpty() bool {
	if nil == hn {
		return true
	}
	for _, m := range defaultMethodEvaluateOrder {
		if n := hn.getHandlerNameByMethod(m); n != "" {
			return false
		}
	}
	return true
}

func (hn *HandlerNames) String() string {
	t, err := json.Marshal(hn)
	if nil != err {
		return err.Error()
	}
	return string(t)
}
