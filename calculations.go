package main

import (
	"strings"
	"strconv"
	"sync"
	"log"
)

func parsePointString(points_str string) ([]int64, []int64) {
	log.Printf("Point string: %v", points_str)
	
	point_strs := strings.Split(points_str, "], [")
	log.Printf("Split point strings: %v", point_strs)

	als_vals := make([]int64, len(point_strs))
	bl_adjs := make([]int64, len(point_strs))
	var wg sync.WaitGroup
	var str_mtx, point_mtx sync.Mutex
	
	for ind := range point_strs {
		wg.Add(1)
		go func(i int) {

			str_mtx.Lock()
			point_strs[i] = strings.TrimSuffix(strings.TrimPrefix(point_strs[i], "[["), "]]\x00")
			str_mtx.Unlock()
			cur_point_strs := strings.Split(point_strs[i], ", ")
			cur_point_bl, err := strconv.ParseInt(cur_point_strs[0], 0, 64)
			check(err)

			cur_point_als, err := strconv.ParseInt(cur_point_strs[1], 0, 64)
			check(err)

			point_mtx.Lock()
			als_vals[i] = cur_point_als
			bl_adjs[i] = cur_point_bl
			point_mtx.Unlock()
			wg.Done()
		}(ind)
	}
	wg.Wait()
	return als_vals, bl_adjs
}

func calcBrightAdj(illum int64, prevNextAls []int64, prevNextBl []int64) int64 {
	if prevNextAls[0] == prevNextAls[1] {
		return prevNextBl[1]
	}
	return int64((float64(illum - prevNextAls[0]) / float64(prevNextAls[1] - prevNextAls[0]))) * (prevNextBl[1] - prevNextBl[0]) + prevNextBl[0];
}
