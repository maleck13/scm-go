package data_test

import "testing"

func getTopLevelRef() string {
	return `
	{
	  "repoUrl":"git:/temp/temp.git",
	  "repoBranch":"refs/heads/master",
	  "repoType":"branch",
	  "bare":false,
	  "repoCommitHash":"dsdasdsadsd0sadas0d90asd9as0",
	  "clusterName":"development",
	  "appGuid":"anapp"	
	}
	
	`
}

func TestRefDetailsAsString(t *testing.T) {

}

func TestRefDetailsAsObject(t *testing.T) {

}

func TestRefDetailsAsTopLevel(t *testing.T) {

}

func TestGetBranchRef(t *testing.T) {

}

func TestGetTagRef(t *testing.T) {

}

func TestGetCommitRef(t *testing.T) {

}
