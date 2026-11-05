package sinks

import (
	"errors"
	"fmt"
	"log/slog"

	"dev.kaesebrot.eu/go/feedrrr/internal/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/utility"
	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

const maxRecursionDepth = 10

func SetupSinks(resultJobSinks *map[string]*router.ServiceRouter, c *config.FeedrrrConfig) error {
	sinkAliases, err := flattenSinks(&c.Sinks)
	if err != nil {
		return err
	}

	for jobName, jobConfig := range c.Jobs {
		resolvedJobSinks := []string{}
		for _, sink := range jobConfig.Sinks {
			if utility.IsUrl(sink) {
				resolvedJobSinks = append(resolvedJobSinks, sink)
			} else {
				sinks, ok := sinkAliases[sink]
				if !ok {
					return fmt.Errorf("Unknown sink alias: %v", sink)
				}
				resolvedJobSinks = append(resolvedJobSinks, sinks...)
			}
		}

		sender, err := shoutrrr.CreateSender(resolvedJobSinks...)
		if err != nil {
			return err
		}

		(*resultJobSinks)[jobName] = sender
	}

	return nil
}

func flattenSinks(sinkAliases *map[string][]string) (map[string][]string, error) {
	flattened := make(map[string][]string)
	for name, _ := range *sinkAliases {
		flattenedAlias, err := flattenSinkAliasRecursively(sinkAliases, name, 0)
		if err != nil {
			return flattened, err
		}
		flattened[name] = flattenedAlias
	}

	return flattened, nil
}

func flattenSinkAliasRecursively(sinkAliases *map[string][]string, name string, currentDepth int) ([]string, error) {
	results := []string{}
	sinks, ok := (*sinkAliases)[name]
	if !ok {
		return results, fmt.Errorf("Unknown sink alias: %v", name)
	}

	for _, sink := range sinks {
		if utility.IsUrl(sink) {
			results = append(results, sink)
		} else {
			if currentDepth >= maxRecursionDepth {
				return results, errors.New("Max recursion depth reached!")
			}
			flattened, err := flattenSinkAliasRecursively(sinkAliases, sink, currentDepth+1)
			if err != nil {
				return results, err
			}
			results = append(results, flattened...)
		}
	}

	slog.Default().Debug("Resolved sink alias", "alias", name, "urls", results)

	return results, nil
}

func SendToSink(sinkMap *map[string]*router.ServiceRouter, sink string, message string, params *types.Params) error {
	(*sinkMap)[sink].Send(message, params)
	return nil
}
