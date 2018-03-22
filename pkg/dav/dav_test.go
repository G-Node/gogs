package dav

import "testing"

func TestGetRepoName(t *testing.T){
	name, err := getRName("https://gin.g-node.org/cgars/test/_dav/adasdasd/daasdas/asdasdsa")
	if err != nil{
		t.Logf("Repo Name not dtermined from path")
		t.Fail()
		return
	}
	if name != "test"{
		t.Logf("Repo Name not dtermined from path")
		t.Fail()
		return
	}
	return
}
