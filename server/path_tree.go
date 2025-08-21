package server

import (
	"fmt"
	"strings"
	"vivalchemy/http-server-from-scratch/request"
	"vivalchemy/http-server-from-scratch/response"
)

type HandlerError error

var (
	HandlerErrorMethodNotAllowed HandlerError = fmt.Errorf("method not allowed")
	HandlerErrorNotFound         HandlerError = fmt.Errorf("url not found")
)

type Handler func(res *response.Writer, req *request.Request)

type PathTreeNode struct {
	section        string
	AllowedMethods map[HTTPMethod]Handler   // METHOD -> Handler
	children       map[string]*PathTreeNode // path section -> Its tree
}

func NewPathTree() *PathTreeNode {
	return &PathTreeNode{
		section:        "",
		AllowedMethods: make(map[HTTPMethod]Handler),
		children:       make(map[string]*PathTreeNode),
	}
}

func (t *PathTreeNode) add(method HTTPMethod, path string, handler Handler) {
	// Handle root path specially
	if path == "/" || path == "" {
		if t.AllowedMethods[method] != nil {
			panic("duplicate handler for method")
		}
		t.AllowedMethods[method] = handler
		return
	}

	// Split path into sections, removing empty segments
	pathSections := strings.Split(strings.Trim(path, "/"), "/")

	// Filter out empty sections
	var cleanSections []string
	for _, section := range pathSections {
		if section != "" {
			cleanSections = append(cleanSections, section)
		}
	}
	pathSections = cleanSections

	currentTree := t
	for i, section := range pathSections {
		if i == len(pathSections)-1 {
			// Last section - create child node and add handler to it
			if _, ok := currentTree.children[section]; !ok {
				currentTree.children[section] = NewPathTree()
				currentTree.children[section].section = section
			}

			// Add handler to the child node
			if currentTree.children[section].AllowedMethods[method] != nil {
				t.print(0)
				panic("duplicate handler for method")
			}
			currentTree.children[section].AllowedMethods[method] = handler
		} else {
			// Intermediate section - create child if doesn't exist
			if _, ok := currentTree.children[section]; !ok {
				currentTree.children[section] = NewPathTree()
				currentTree.children[section].section = section
			}
			currentTree = currentTree.children[section]
		}
	}
}

func (t *PathTreeNode) find(method HTTPMethod, path string) (Handler, error) {
	pathSections := strings.Split(strings.Trim(path, "/"), "/")
	// for / it will return an empty array
	currentTree := t

	for _, section := range pathSections {
		if child, ok := currentTree.children[section]; ok {
			// it has a child traverse to that child
			currentTree = child
		} else if wildcard, ok := currentTree.children["*"]; ok {
			currentTree = wildcard
			break
		} else {
			return nil, HandlerErrorNotFound
		}
	}

	if handler, ok := currentTree.AllowedMethods[method]; ok {
		// match the handler directly
		// doesn't resolve the /path if the route mentioned is /path/*
		// so need the else if part to do it
		return handler, nil
	} else if wildcard, ok := currentTree.children["*"]; ok {
		// if there is /domain/* then it will match /domain and /domain/anything
		return wildcard.AllowedMethods[method], nil
	}
	return nil, HandlerErrorMethodNotAllowed
}

func (t *PathTreeNode) addOptions() {
	if len(t.AllowedMethods) > 0 {
		t.AllowedMethods[MethodOptions] = Handler(func(res *response.Writer, req *request.Request) {
			res.WriteStatusLine(response.StatusOk)
			headers := response.GetDefaultHeaders(0)
			headers.Delete("Content-Type")

			// Collect all allowed methods
			var methods []string
			for method := range t.AllowedMethods {
				methods = append(methods, string(method))
			}

			// Set Allow header with comma-separated methods
			headers.Set("Allow", strings.Join(methods, ", "))
			res.WriteHeaders(*headers)
		})
	}

	// Recursively add OPTIONS to all children
	for _, child := range t.children {
		child.addOptions()
	}
}

func (t *PathTreeNode) print(depth int) {
	indent := strings.Repeat("  ", depth)

	if t.section != "" {
		fmt.Printf("%sSection: %s\n", indent, t.section)
	}

	if len(t.AllowedMethods) > 0 {
		fmt.Printf("%sAllowedMethods:\n", indent)
		for method := range t.AllowedMethods {
			fmt.Printf("%s  - %v\n", indent, method)
		}
	}

	if len(t.children) > 0 {
		fmt.Printf("%sChildren:\n", indent)
		for section, child := range t.children {
			fmt.Printf("%s  %s:\n", indent, section)
			child.print(depth + 2) // Fixed: call print on the child, not t
		}
	}
}
