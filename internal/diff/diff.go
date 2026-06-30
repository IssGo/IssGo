// Package diff provides simple text differencing.
package diff

import (
	"fmt"
	"strings"
)

type Line struct {
	Kind rune   // '+', '-', ' '
	Text string
	NumA int
	NumB int
}

func Compute(a, b string) []Line {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")

	// Simple LCS-based diff
	m, n := len(aLines), len(bLines)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if aLines[i-1] == bLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var result []Line
	i, j := m, n
	var rev []Line

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && aLines[i-1] == bLines[j-1] {
			rev = append(rev, Line{Kind: ' ', Text: aLines[i-1], NumA: i, NumB: j})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			rev = append(rev, Line{Kind: '+', Text: bLines[j-1], NumB: j})
			j--
		} else {
			rev = append(rev, Line{Kind: '-', Text: aLines[i-1], NumA: i})
			i--
		}
	}

	for k := len(rev) - 1; k >= 0; k-- {
		result = append(result, rev[k])
	}
	return result
}

func Format(lines []Line) string {
	var sb strings.Builder
	for _, l := range lines {
		fmt.Fprintf(&sb, "%c %s\n", l.Kind, l.Text)
	}
	return sb.String()
}
