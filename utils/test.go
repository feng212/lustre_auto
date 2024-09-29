package main

// 回文串
// s[i,j]是不是回文串
//
//		 1.s[i]==s[j]
//		  1.1. i-j 0或1 无需判断
//	   1.2  s[i+1,j-1]==true
//	   1.3  s[i,j-2]
//
// dp[i,j]表示是否是一个回文串
//
//	dp[i,j] :  (s[i]==s[j],(dp[i+1,j-1],dp[i,j-2]))
func longestPalindromeSubseq(s string) int {
	dp := make([][]int, len(s))
	for i := 0; i < len(s); i++ {
		dp[i] = make([]int, len(s))
		dp[i][i] = 1
	}
	for j := 1; j < len(s); j++ {
		for i := 0; i < j; i++ {
			if s[i] == s[j] {
				if j-i <= 1 {
					dp[i][j] = j - i + 1
				} else {
					dp[i][j] = dp[i+1][j-1] + 2
				}

			} else {
				dp[i][j] = max(dp[i][j-1], dp[i+1][j])
			}
		}
	}
	return dp[0][len(s)-1]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func main() {
	longestPalindromeSubseq("cbbd")
}
