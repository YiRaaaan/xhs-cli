package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Suoyiran1/xhs-cli/internal/errors"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

type SearchResult struct {
	Search struct {
		Feeds FeedsValue `json:"feeds"`
	} `json:"search"`
}

type FilterOption struct {
	SortBy      string `json:"sort_by,omitempty"`
	NoteType    string `json:"note_type,omitempty"`
	PublishTime string `json:"publish_time,omitempty"`
	SearchScope string `json:"search_scope,omitempty"`
	Location    string `json:"location,omitempty"`
}

type internalFilterOption struct {
	FiltersIndex int
	TagsIndex    int
	Text         string
}

var filterOptionsMap = map[int][]internalFilterOption{
	1: {
		{FiltersIndex: 1, TagsIndex: 1, Text: "综合"},
		{FiltersIndex: 1, TagsIndex: 2, Text: "最新"},
		{FiltersIndex: 1, TagsIndex: 3, Text: "最多点赞"},
		{FiltersIndex: 1, TagsIndex: 4, Text: "最多评论"},
		{FiltersIndex: 1, TagsIndex: 5, Text: "最多收藏"},
	},
	2: {
		{FiltersIndex: 2, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 2, TagsIndex: 2, Text: "视频"},
		{FiltersIndex: 2, TagsIndex: 3, Text: "图文"},
	},
	3: {
		{FiltersIndex: 3, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 3, TagsIndex: 2, Text: "一天内"},
		{FiltersIndex: 3, TagsIndex: 3, Text: "一周内"},
		{FiltersIndex: 3, TagsIndex: 4, Text: "半年内"},
	},
	4: {
		{FiltersIndex: 4, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 4, TagsIndex: 2, Text: "已看过"},
		{FiltersIndex: 4, TagsIndex: 3, Text: "未看过"},
		{FiltersIndex: 4, TagsIndex: 4, Text: "已关注"},
	},
	5: {
		{FiltersIndex: 5, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 5, TagsIndex: 2, Text: "同城"},
		{FiltersIndex: 5, TagsIndex: 3, Text: "附近"},
	},
}

func convertToInternalFilters(filter FilterOption) ([]internalFilterOption, error) {
	var internalFilters []internalFilterOption

	if filter.SortBy != "" {
		internal, err := findInternalOption(1, filter.SortBy)
		if err != nil {
			return nil, fmt.Errorf("排序依据错误: %w", err)
		}
		internalFilters = append(internalFilters, internal)
	}
	if filter.NoteType != "" {
		internal, err := findInternalOption(2, filter.NoteType)
		if err != nil {
			return nil, fmt.Errorf("笔记类型错误: %w", err)
		}
		internalFilters = append(internalFilters, internal)
	}
	if filter.PublishTime != "" {
		internal, err := findInternalOption(3, filter.PublishTime)
		if err != nil {
			return nil, fmt.Errorf("发布时间错误: %w", err)
		}
		internalFilters = append(internalFilters, internal)
	}
	if filter.SearchScope != "" {
		internal, err := findInternalOption(4, filter.SearchScope)
		if err != nil {
			return nil, fmt.Errorf("搜索范围错误: %w", err)
		}
		internalFilters = append(internalFilters, internal)
	}
	if filter.Location != "" {
		internal, err := findInternalOption(5, filter.Location)
		if err != nil {
			return nil, fmt.Errorf("位置距离错误: %w", err)
		}
		internalFilters = append(internalFilters, internal)
	}

	return internalFilters, nil
}

func findInternalOption(filtersIndex int, text string) (internalFilterOption, error) {
	options, exists := filterOptionsMap[filtersIndex]
	if !exists {
		return internalFilterOption{}, fmt.Errorf("筛选组 %d 不存在", filtersIndex)
	}
	for _, option := range options {
		if option.Text == text {
			return option, nil
		}
	}
	return internalFilterOption{}, fmt.Errorf("在筛选组 %d 中未找到文本 '%s'", filtersIndex, text)
}

func validateInternalFilterOption(filter internalFilterOption) error {
	if filter.FiltersIndex < 1 || filter.FiltersIndex > 5 {
		return fmt.Errorf("无效的筛选组索引 %d，有效范围为 1-5", filter.FiltersIndex)
	}
	options, exists := filterOptionsMap[filter.FiltersIndex]
	if !exists {
		return fmt.Errorf("筛选组 %d 不存在", filter.FiltersIndex)
	}
	if filter.TagsIndex < 1 || filter.TagsIndex > len(options) {
		return fmt.Errorf("筛选组 %d 的标签索引 %d 超出范围，有效范围为 1-%d",
			filter.FiltersIndex, filter.TagsIndex, len(options))
	}
	return nil
}

type SearchAction struct {
	page *rod.Page
}

func NewSearchAction(page *rod.Page) *SearchAction {
	pp := page.Timeout(60 * time.Second)
	return &SearchAction{page: pp}
}

func (s *SearchAction) Search(ctx context.Context, keyword string, filters ...FilterOption) ([]Feed, error) {
	page := s.page.Context(ctx)

	searchURL := makeSearchURL(keyword)
	page.MustNavigate(searchURL)
	page.MustWaitStable()

	page.MustWait(`() => window.__INITIAL_STATE__ !== undefined`)

	if len(filters) > 0 {
		// Collect filter text values to click
		var filterTexts []string
		for _, filter := range filters {
			if filter.SortBy != "" {
				filterTexts = append(filterTexts, filter.SortBy)
			}
			if filter.NoteType != "" {
				filterTexts = append(filterTexts, filter.NoteType)
			}
			if filter.PublishTime != "" {
				filterTexts = append(filterTexts, filter.PublishTime)
			}
			if filter.SearchScope != "" {
				filterTexts = append(filterTexts, filter.SearchScope)
			}
			if filter.Location != "" {
				filterTexts = append(filterTexts, filter.Location)
			}
		}

		if len(filterTexts) > 0 {
			if err := applyFilters(page, filterTexts); err != nil {
				logrus.Warnf("筛选应用失败，返回默认结果: %v", err)
			} else {
				page.MustWaitStable()
				page.MustWait(`() => window.__INITIAL_STATE__ !== undefined`)
			}
		}
	}

	result := page.MustEval(`() => {
		if (window.__INITIAL_STATE__ &&
		    window.__INITIAL_STATE__.search &&
		    window.__INITIAL_STATE__.search.feeds) {
			const feeds = window.__INITIAL_STATE__.search.feeds;
			const feedsData = feeds.value !== undefined ? feeds.value : feeds._value;
			if (feedsData) {
				return JSON.stringify(feedsData);
			}
		}
		return "";
	}`).String()

	if result == "" {
		return nil, errors.ErrNoFeeds
	}

	var feeds []Feed
	if err := json.Unmarshal([]byte(result), &feeds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feeds: %w", err)
	}

	return feeds, nil
}

// applyFilters opens the filter panel and clicks options by text content.
func applyFilters(page *rod.Page, filterTexts []string) error {
	// Click the filter button to open the panel
	filterBtn, err := page.Timeout(10 * time.Second).Element(`div.filter`)
	if err != nil {
		return fmt.Errorf("未找到筛选按钮: %w", err)
	}
	filterBtn.MustClick()
	time.Sleep(1 * time.Second)

	// Wait for filter panel with retry
	for i := 0; i < 3; i++ {
		has, _, _ := page.Has(`.filter-panel`)
		if has {
			break
		}
		filterBtn.MustClick()
		time.Sleep(1 * time.Second)
	}

	// Use JavaScript to click filter options by matching text content
	for _, text := range filterTexts {
		clicked := page.MustEval(`(targetText) => {
			const panel = document.querySelector('.filter-panel');
			if (!panel) return false;
			// Find all clickable elements in the panel
			const allEls = panel.querySelectorAll('.tag, .tags, span, div');
			for (const el of allEls) {
				if (el.textContent.trim() === targetText && el.offsetParent !== null) {
					el.click();
					return true;
				}
			}
			return false;
		}`, text).Bool()

		if clicked {
			logrus.Debugf("筛选: 点击了 '%s'", text)
			time.Sleep(500 * time.Millisecond)
		} else {
			logrus.Warnf("筛选: 未找到 '%s'", text)
		}
	}

	time.Sleep(1 * time.Second)
	return nil
}

func makeSearchURL(keyword string) string {
	values := url.Values{}
	values.Set("keyword", keyword)
	values.Set("source", "web_explore_feed")
	return fmt.Sprintf("https://www.xiaohongshu.com/search_result?%s", values.Encode())
}
