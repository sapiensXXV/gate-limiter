package metrics

import (
	"path"
	"regexp"
	"strings"
)

var (
	reInt   = regexp.MustCompile(`^\d+$`)
	reUUID  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	reHex24 = regexp.MustCompile(`^[0-9a-fA-F]{24}$`) // MongoDB ObjectID 등
)

// 쿼리 스트링 제거 -> 중복 슬래시 정리 -> 각 세그먼트에서 숫자/UUID/ObjectID를 표준 토큰으로 치환한다.
// 라벨 카디널리티 폭발 방지를 위해서 사용된다.
// /api/item/123/comment와 /api/item/321/comment를 같은 라벨로 묶는다.
func normalizeMetricName(p string) string {
	if p == "" || p == "/" {
		return "/"
	}

	// 1. 쿼리스트링 제거
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}

	// 2. /./ // 같은 것 정리
	p = path.Clean(p)

	parts := strings.Split(p, "/")
	out := make([]string, 0, len(parts))
	for _, seg := range parts {
		if seg == "" {
			continue
		}
		switch {
		case reInt.MatchString(seg):
			out = append(out, ":id")
		case reUUID.MatchString(seg):
			out = append(out, ":uuid")
		case reHex24.MatchString(seg):
			out = append(out, ":hex24")
		default:
			out = append(out, seg)
		}
	}

	return "/" + strings.Join(out, "/")

}
