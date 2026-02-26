package router

import (
	"sort"
	"strings"
)

type RouteConfig struct {
	EventList []string
	GroupList []string
	Enabled   bool
	Priority  int
}

type groupItem struct {
	Name     string
	Priority int
}

type RouteMap map[string][]string

func New(configs []RouteConfig) RouteMap {
	tempMap := make(map[string][]groupItem, len(configs)*2)

	for _, config := range configs {
		if !config.Enabled {
			// 配置项支持开关能力
			continue
		}

		for _, event := range config.EventList {
			for _, group := range config.GroupList {
				tempMap[event] = append(tempMap[event], groupItem{Name: parseGroupName(event, group), Priority: config.Priority})
			}
		}
	}

	routeMap := make(RouteMap, len(tempMap))
	for event, groupItems := range tempMap {
		// 按照优先级排序, 优先级数值越大优先执行
		sort.Slice(groupItems, func(i, j int) bool {
			return groupItems[i].Priority > groupItems[j].Priority
		})

		groupNames := make([]string, len(groupItems))
		for i, item := range groupItems {
			groupNames[i] = item.Name
		}
		routeMap[event] = groupNames
	}

	return routeMap
}

// 解析分组名称, 支持动态参数, 例如: $符号表示当前Event
func parseGroupName(event, group string) string {
	return strings.ReplaceAll(group, "$", event)
}
