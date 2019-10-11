package gourmet

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"plugin"

	mapset "github.com/deckarep/golang-set"
)

var (
	registeredAnalyzers []Analyzer
	resolvedGraph       analyzerGraph
)

type Result interface {
	Key() string
}

type Analyzer interface {
	Filter(c *Connection) bool
	Analyze(c *Connection) (Result, error)
}

// This function needs some major refactoring...
func newAnalyzers(links map[string]interface{}, skipUpdate bool) (err error) {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeDir := usr.HomeDir
	pluginsDir := filepath.Join(homeDir, ".gourmet/plugins/")
	var analyzerFiles []string
	for _, analyzer := range resolvedGraph {
		pluginDir := filepath.Join(pluginsDir, analyzer.name)
		mainPath := filepath.Join(pluginDir, "main.go")
		exists, err := dirExists(pluginDir)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Printf("[*] Installing %s\n", analyzer.name)
			err = exec.Command("git", "clone", fmt.Sprintf("https://%s", analyzer.name), pluginDir).Run()
			if err != nil {
				return fmt.Errorf("failed to install %s: %s", analyzer.name, err.Error())
			}
		} else if !skipUpdate {
			fmt.Printf("[*] Updating %s\n", analyzer.name)
			err = exec.Command("git", "-C", pluginDir, "pull").Run()
		}
		_, err = os.Stat(mainPath)
		if err != nil {
			return err
		}
		analyzerFiles = append(analyzerFiles, mainPath)
		setAnalyzerConfig(analyzer.name, links[analyzer.name])
	}
	if len(analyzerFiles) > 0 {
		for _, analyzerFile := range analyzerFiles {
			folderName := filepath.Dir(analyzerFile)
			fmt.Printf("[*] Building %s\n", filepath.Base(filepath.Dir(analyzerFile)))
			out, err := exec.Command("go", "build", "-buildmode=plugin", "-o",
				fmt.Sprintf("%s/main.so", filepath.Dir(analyzerFile)), analyzerFile).CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to build %s: %s", analyzerFile, string(out))
			}
			p, err := plugin.Open(fmt.Sprintf("%s/main.so", folderName))
			if err != nil {
				return err
			}
			newAnalyzerFunc, err := p.Lookup("NewAnalyzer")
			if err != nil {
				return fmt.Errorf("Failed lookup of NewAnalyzer in %s: %s", analyzerFile, err.Error())
			}
			analyzerFunc, ok := newAnalyzerFunc.(func() Analyzer)
			if !ok {
				return fmt.Errorf("NewAnalyzer in %s does not return an Analyzer interface", analyzerFile)
			}
			analyzer := analyzerFunc()
			registeredAnalyzers = append(registeredAnalyzers, analyzer)
		}
	}
	return nil
}

func createAnalyzerNode(name string, config interface{}) (*node, error) {
	// check if analyzer has any arguments
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return &node{
			name: name,
			deps: nil,
		}, nil
	}
	// check if analyzer has depends_on argument
	dependencies, ok := configMap["depends_on"]
	if !ok {
		return &node{
			name: name,
			deps: nil,
		}, nil
	}
	// if depends_on exists, make sure it is a list
	depList, ok := dependencies.([]interface{})
	if !ok {
		return nil, fmt.Errorf("depends_on for %s is not a list", name)
	}
	var deps []string
	for _, dep := range depList {
		// for each element of depends_on, make sure it is a string
		depString, ok := dep.(string)
		if !ok {
			return nil, fmt.Errorf("depends_on list value for %s is not a string", name)
		}
		deps = append(deps, depString)
	}
	return &node{
		name: name,
		deps: deps,
	}, nil
}

// Copyright (c) 2016 Marin Atanasov Nikolov <dnaeon@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer
//    in this position and unchanged.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR(S) ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHOR(S) BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Node represents a single node in the graph with it's dependencies
type node struct {
	// Name of the node
	name string
	// Dependencies of the node
	deps []string
}

type analyzerGraph []*node

// Resolves the dependency graph
func resolveGraph(graph analyzerGraph) error {
	// A map containing the node names and the actual node object
	nodeNames := make(map[string]*node)
	// A map containing the nodes and their dependencies
	nodeDependencies := make(map[string]mapset.Set)
	// Populate the maps
	for _, node := range graph {
		nodeNames[node.name] = node
		dependencySet := mapset.NewSet()
		for _, dep := range node.deps {
			dependencySet.Add(dep)
		}
		nodeDependencies[node.name] = dependencySet
	}
	// Iteratively find and remove nodes from the graph which have no dependencies.
	// If at some point there are still nodes in the graph and we cannot find
	// nodes without dependencies, that means we have a circular dependency
	var resolved analyzerGraph
	for len(nodeDependencies) != 0 {
		// Get all nodes from the graph which have no dependencies
		readySet := mapset.NewSet()
		for name, deps := range nodeDependencies {
			if deps.Cardinality() == 0 {
				readySet.Add(name)
			}
		}
		// If there aren't any ready nodes, then we have a cicular dependency
		if readySet.Cardinality() == 0 {
			var g analyzerGraph
			for name := range nodeDependencies {
				g = append(g, nodeNames[name])
			}
			return errors.New("circular dependency or missing dependency found")
		}
		// Remove the ready nodes and add them to the resolved graph
		for name := range readySet.Iter() {
			delete(nodeDependencies, name.(string))
			resolved = append(resolved, nodeNames[name.(string)])
		}
		// Also make sure to remove the ready nodes from the
		// remaining node dependencies as well
		for name, deps := range nodeDependencies {
			diff := deps.Difference(readySet)
			nodeDependencies[name] = diff
		}
	}
	resolvedGraph = resolved
	return nil
}
