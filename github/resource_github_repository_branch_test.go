package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v28/github"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGithubBranch_basic(t *testing.T) {
	var reference github.Reference

	repoName := "tf-acc-test-repo-branch-" + acctest.RandString(5)
	branchName := "foobar"
	resourceName := "github_repository_branch." + branchName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccGithubRepositoryBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubRepositoryBranchConfig(repoName, branchName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubRepositoryBranchExists(
						resourceName, repoName+":"+branchName, &reference),
					resource.TestCheckResourceAttr(resourceName, "repository", repoName),
					resource.TestCheckResourceAttr(resourceName, "branch", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "source_branch", "master"),
				),
			},
			{
				Config: testAccGithubRepositoryBranchConfig(repoName, branchName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubRepositoryBranchExists(
						resourceName, repoName+":"+branchName, &reference),
					resource.TestCheckResourceAttr(resourceName, "repository", repoName),
					resource.TestCheckResourceAttr(resourceName, "branch", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "source_branch", "master"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"sha",
					"source_branch",
				},
			},
		},
	})
}

func testAccGithubRepositoryBranchConfig(repoName, branchName string) string {
	return fmt.Sprintf(`
resource "github_repository" "%s" {
  name        = "%s"
  description = "Terraform Acceptance Test"
  auto_init   = true
}

resource "github_repository_branch" "%s" {
  repository = github_repository.%s.name
  branch     = "%s"
}
`, repoName, repoName, branchName, repoName, branchName)
}

func testAccGithubRepositoryBranchDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_repository_branch" {
			continue
		}

		conn := testAccProvider.Meta().(*Organization).client
		orgName := testAccProvider.Meta().(*Organization).name
		repoName, branchName, err := parseTwoPartID(rs.Primary.ID, "repository", "branch")
		if err != nil {
			return err
		}

		ref, resp, err := conn.Git.GetRef(context.TODO(), orgName, repoName, branchName)
		if err == nil {
			if ref != nil {
				return fmt.Errorf("Repository branch still exists: %s/%s (%s)",
					orgName, repoName, branchName)
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckGithubRepositoryBranchExists(n, id string, reference *github.Reference) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID != id {
			return fmt.Errorf("Expected ID to be %v, got %v", id, rs.Primary.ID)
		}

		conn := testAccProvider.Meta().(*Organization).client
		orgName := testAccProvider.Meta().(*Organization).name
		repoName, branchName, err := parseTwoPartID(rs.Primary.ID, "repository", "branch")
		if err != nil {
			return err
		}

		ref, _, err := conn.Git.GetRef(context.TODO(), orgName, repoName, "refs/heads/"+branchName)
		if err != nil {
			return err
		}

		*reference = *ref
		return nil
	}
}
