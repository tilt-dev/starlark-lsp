## Issues
def create_issue(owner: str, repo: str, title: str, body: str, assignee: str, milestone: str, labels: str, assignees: str):
    """Create an issue


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/issues#create-an-issue
    """
    pass

def get_issue(owner: str, repo: str, number: str):
    """Get an issue


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/issues#get-an-issue
    """
    pass

def update_issue(owner: str, repo: str, number: str, title: str, body: str, assignee: str, state: str, stateReason: str, milestone: str, labels: str, assignees: str):
    """Update an issue


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/issues#update-an-issue
    """
    pass

def list_repository_issues(owner: str, repo: str, milestone: str, state: str, assignee: str, creator: str, mentioned: str, labels: str, sort: str, direction: str, since: str):
    """List repository issues


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/issues#list-repository-issues
    """
    pass


## Issue comments
def create_issue_comment(owner: str, repo: str, number: str, body: str):
    """Create an issue comment


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/comments#create-an-issue-comment
    """
    pass

# def get_issue_comment():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def update_issue_comment():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def list_issue_comments():
#      """
#      []string{owner: str, repo: str, number: str},


## Issue labels
def add_issue_labels(owner: str, repo: str, number: str, labels: str):
    """Add labes to an issue


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/labels#add-labels-to-an-issue
    """
    pass

def remove_issue_label(owner: str, repo: str, number: str, label: str):
    """Remove a label from an issue


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/issues/labels#remove-a-label-from-an-issue
    """
    pass

# def set_issue_labels():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def remove_all_issue_labels():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def list_issue_labels():
#      """
#      []string{owner: str, repo: str, number: str},


## Pull requests.
def get_pull_request(owner: str, repo: str, number: str):
    """Get a pull request


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request
    """
    pass

def list_pull_requests(owner: str, repo: str, state: str, head: str, base: str, sort: str, direction: str):
    """List pull requests


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/pulls/pulls#list-pull-requests
    """
    pass

def create_pull_request(owner: str, repo: str, head: str, base: str, title: str|None, body: str|None, head_repo: str|None, draft: str|None, issue: str|None, maintainer_can_modify: str|None):
    """Create a pull request


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request
    """
    pass

## Pull-request comments.
def list_review_comments(owner: str, repo: str, number: str):
    """List review comments on a pull request


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/pulls/comments#list-review-comments-on-a-pull-request
    """
    pass

# def create_review_comment():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def get_review_comment():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def update_review_comment():
#      """
#      []string{owner: str, repo: str, number: str},
#
# def create_review_comment_reply():
#      """
#      []string{owner: str, repo: str, number: str},


## Reactions.
def create_reaction_for_commit_comment(owner: str, repo: str, id: str, content: str):
    """Creates reaction for a commit comment


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-commit-comment
    """
    pass

def create_reaction_for_issue(owner: str, repo: str, number: str, content: str):
    """Creates reaction for an issue


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue"""
    pass

def create_reaction_for_issue_comment(owner: str, repo: str, id: str, content: str):
    """Create reaction for an issue comment


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue-comment
    """
    pass

def create_reaction_for_pull_request_review_comment(owner: str, repo: str, id: str, content: str):
    """Create reaction for a pull request review comment


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-pull-request-review-comment
    """
    pass


## Repository Contents.
def create_file(owner: str, repo: str, path: str, content: str, message: str, sha: str|None, branch: str|None, committer: str|None):
    """Create or update file contents


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/repos/contents#create-or-update-file-contents
    """
    pass

def get_contents(owner: str, repo: str, path: str, ref: str|None):
    """Get repository content


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/repos/contents#get-repository-content
    """
    pass

## Git references.
def create_ref(owner: str, repo: str, ref: str, sha: str):
    """Create a reference


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/git/refs#create-a-reference
    """
    pass

def get_ref(owner: str, repo: str, ref: str):
    """Get a reference


    Args:


    Returns:


    API:
      see https://docs.github.com/en/rest/git/refs#get-a-reference
    """
    pass
