package main

import "fmt"

func findAnagrams(s string, p string) []int {
	// 枚举起点,遍历终点
	countP := [26]int{}
	countS := [26]int{}

	if len(s) < len(p) {
		return []int{}
	}

	for i := range p {
		countP[int(p[i]-'a')]++
		countS[int(s[i]-'a')]++
	}
	fmt.Println(countP)

	//countS[int(s[0]-'a')]++
	countS[int(s[len(p)-1]-'a')]--

	res := []int{}
	for i := len(p) - 1; i < len(s); i++ {

		countS[int(s[i]-'a')]++
		//fmt.Println(countS)
		if countP == countS {
			res = append(res, i-len(p)+1)
		}
		countS[int(s[i-len(p)+1]-'a')]--
	}
	return res

}

func main() {
	findAnagrams("cbaebabacd", "abc")
}
