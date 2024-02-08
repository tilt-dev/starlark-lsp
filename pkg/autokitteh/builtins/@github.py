class Timestamp:
    pass

class GHWorkflowRuns:
    total_count: int
    workflow_runs: List[GHWorkflowRun]

class GHWorkflowRun:
    id: int
    name: str
    node_id: str
    head_branch: str
    head_sha: str
    run_number: int
    run_attempt: int
    event: str
    display_title: str
    status: str
    conclusion: str
    workflow_id: int
    check_suite_id: int
    check_suite_node_id: str
    url: str
    html_url: str
    pull_requests: List[GHPullRequest]
    created_at: Timestamp
    updated_at: Timestamp
    run_started_at: Timestamp
    jobs_url: str
    logs_url: str
    check_suite_url: str
    artifacts_url: str
    cancel_url: str
    rerun_url: str
    previous_attempt_url: str
    head_commit: GHHeadCommit
    workflow_url: str
    repository: GHRepository
    head_repository: GHRepository
    actor: GHUser
    triggering_actor: GHUser

class GHRepository:
    id: int
    node_id: str
    owner: GHUser
    name: str
    full_name: str
    description: str
    homepage: str
    code_of_conduct: GHCodeOfConduct
    default_branch: str
    master_branch: str
    created_at: Timestamp
    pushed_at: Timestamp
    updated_at: Timestamp
    html_url: str
    clone_url: str
    git_url: str
    mirror_url: str
    ssh_url: str
    svn_url: str
    language: str
    fork: bool
    forks_count: int
    network_count: int
    open_issues_count: int
    open_issues: int
    stargazers_count: int
    subscribers_count: int
    watchers_count: int
    watchers: int
    size: int
    auto_init: bool
    parent: GHRepository
    source: GHRepository
    template_repository: GHRepository
    organization: GHOrganization
    permissions: Dict[str, bool]
    allow_rebase_merge: bool
    allow_update_branch: bool
    allow_squash_merge: bool
    allow_merge_commit: bool
    allow_auto_merge: bool
    allow_forking: bool
    web_commit_signoff_required: bool
    delete_branch_on_merge: bool
    use_squash_pr_title_as_default: bool
    squash_merge_commit_title: str
    squash_merge_commit_message: str
    merge_commit_title: str
    merge_commit_message: str
    topics: List[str]
    archived: bool
    disabled: bool

    license: GHLicense

    private: bool
    has_issues: bool
    has_wiki: bool
    has_pages: bool
    has_projects: bool
    has_downloads: bool
    has_discussions: bool
    is_template: bool
    license_template: str
    gitignore_template: str

    security_and_analysis: GHSecurityAndAnalysis

    team_id: int
    url: str
    archive_url: str
    assignees_url: str
    blobs_url: str
    branches_url: str
    collaborators_url: str
    comments_url: str
    commits_url: str
    compare_url: str
    contents_url: str
    contributors_url: str
    deployments_url: str
    downloads_url: str
    events_url: str
    forks_url: str
    git_commits_url: str
    git_refs_url: str
    git_tags_url: str
    hooks_url: str
    issue_comment_url: str
    issue_events_url: str
    issues_url: str
    keys_url: str
    labels_url: str
    languages_url: str
    merges_url: str
    milestones_url: str
    notifications_url: str
    pulls_url: str
    releases_url: str
    stargazers_url: str
    statuses_url: str
    subscribers_url: str
    subscription_url: str
    tags_url: str
    trees_url: str
    teams_url: str

    text_matches: List[GHTextMatch]
    visibility: str

    role_name: str

# GHPullRequest represents a GitHub pull request on a repository.
class GHPullRequest:
    id: int
    number: int
    state: str
    locked: bool
    title: str
    body: str
    created_at: Timestamp
    updated_at: Timestamp
    closed_at: Timestamp
    merged_at: Timestamp
    labels: List[GHLabel]
    user: GHUser
    draft: bool
    merged: bool
    mergeable: bool
    mergeable_state: str
    merged_by: GHUser
    merge_commit_sha: str
    rebaseable: bool
    comments: int
    commits: int
    additions: int
    deletions: int
    changed_files: int
    url: str
    html_url: str
    issue_url: str
    statuses_url: str
    diff_url: str
    patch_url: str
    commits_url: str
    comments_url: str
    review_comments_url: str
    review_comment_url: str
    review_comments: int
    assignee: GHUser
    assignees: List[GHUser]
    milestone: GHMilestone
    maintainer_can_modify: bool
    author_association: str
    node_id: str
    requested_reviewers: List[GHUser]
    auto_merge: GHPullRequestAutoMerge

    requested_teams: List[GHTeam]

    links: GHPRLinks
    head: GHPullRequestBranch
    base: GHPullRequestBranch

    active_lock_reason: str


class Organization:
    login: str
    id: int
    node_id: str
    avatar_url: str
    html_url: str
    name: str
    company: str
    blog: str
    location: str
    email: str
    twitter_username: str
    description: str
    public_repos: int
    public_gists: int
    followers: int
    following: int
    created_at: Timestamp
    updated_at: Timestamp
    total_private_repos: int
    owned_private_repos: int
    private_gists: int
    disk_usage: int
    collaborators: int
    billing_email: str
    type: str
    plan: GHPlan
    two_factor_requirement_enabled: bool
    is_verified: bool
    has_organization_projects: bool
    has_repository_projects: bool

    default_repo_permission: str
    default_repo_settings: str

    members_can_create_repos: bool

    members_can_create_public_repos: bool
    members_can_create_private_repos: bool
    members_can_create_internal_repos: bool

    members_can_fork_private_repos: bool

    members_allowed_repository_creation_type: str

    members_can_create_pages: bool
    members_can_create_public_pages: bool
    members_can_create_private_pages: bool
    web_commit_signoff_required: bool
    advanced_security_enabled_for_new_repos: bool
    dependabot_alerts_enabled_for_new_repos: bool
    dependabot_security_updates_enabled_for_new_repos: bool
    dependency_graph_enabled_for_new_repos: bool
    secret_scanning_enabled_for_new_repos: bool
    secret_scanning_push_protection_enabled_for_new_repos: bool

    url: str
    events_url: str
    hooks_url: str
    issues_url: str
    members_url: str
    public_members_url: str
    repos_url: str

class GHUser:
    Login: str
    ID: int
    NodeID: str
    AvatarURL: str
    HTMLURL: str
    GravatarID: str
    Name: str
    Company: str
    Blog: str
    Location: str
    Email: str
    Hireable: bool
    Bio: str
    TwitterUsername: str
    PublicRepos: int
    PublicGists: int
    Followers: int
    Following: int
    CreatedAt: Timestamp
    UpdatedAt: Timestamp
    SuspendedAt: Timestamp
    Type: str
    SiteAdmin: bool
    TotalPrivateRepos: int
    OwnedPrivateRepos: int
    PrivateGists: int
    DiskUsage: int
    Collaborators: int
    TwoFactorAuthentication: bool
    Plan: GHPlan
    LdapDn: str

    URL: str
    EventsURL: str
    FollowingURL: str
    FollowersURL: str
    GistsURL: str
    OrganizationsURL: str
    ReceivedEventsURL: str
    ReposURL: str
    StarredURL: str
    SubscriptionsURL: str

    TextMatches: List[GHTextMatch]

    Permissions: Dict[str, bool]
    RoleName: str

class GHHeadCommit:
    message: str
    author: GHCommitAuthor
    url: str
    distinct: bool

    sha: str

    id: str
    tree_id: str
    timestamp: Timestamp
    committer: GHCommitAuthor
    added: List[str]
    removed: List[str]
    modified: List[str]

class GHCommitAuthor:
    date: Timestamp
    name: str
    email: str

    login: str

class GHCodeOfConduct:
    name: str
    key: str
    url: str
    body: str

class GHPlan:
    name: str
    space: int
    collaborators: int
    private_repos: int
    filled_seats: int
    seats: int

class GHLabel:
    id: int
    url: str
    name: str
    color: str
    description: str
    default: bool
    node_id: str

# Match represents a single text match.
class GHMatch:
    text: str
    indices: List[int]

# TextMatch represents a text match for a SearchResult
class GHTextMatch:
    object_url: str
    object_type: str
    property: str
    fragment: str
    matches: List[GHMatch]

# Milestone represents a GitHub repository milestone.
class GHMilestone:
    url: str
    html_url: str
    labels_url: str
    id: int
    number: int
    state: str
    title: str
    description: str
    creator: GHUser
    open_issues: int
    closed_issues: int
    created_at: Timestamp
    updated_at: Timestamp
    closed_at: Timestamp
    due_on: Timestamp
    node_id: str

class GHRepositoryCommit:
    node_id: str
    sha: str
    commit: GHCommit
    author: GHUser
    committer: User
    parents: List[GHCommit]
    html_url: str
    url: str
    comments_url: str

    stats: GHCommitStats

    files: List[GHCommitFile]


class GHCommit:
    sha: str
    author: GHCommitAuthor
    committer: GHCommitAuthor
    message: str
    tree: GHTree
    parents: List[GHCommit]
    stats: GHCommitStats
    html_url: str
    url: str
    verification: GHSignatureVerification
    node_id: str

    comment_count: int

    signing_key: openpgpEntity

class GHReference:
    ref: str
    url: str
    object: GHGitObject
    node_id: str

class GHGitObject:
    type: str
    sha: str
    url: str

class GHIssueComment:
    id: int
    node_id: str
    body: str
    user: GHUser
    reactions: GHReactions
    created_at: Timestamp
    updated_at: Timestamp

    author_association: str
    url: str
    html_url: str
    issue_url: str

class GHIssue:
    id: int
    number: int
    state: str
    state_reason: str
    locked: bool
    title: str
    body: str
    author_association: str
    user: GHUser
    labels: List[GHLabel]
    assignee: GHUser
    comments: int
    closed_at: Timestamp
    created_at: Timestamp
    updated_at: Timestamp
    closed_by: GHUser
    url: str
    html_url: str
    comments_url: str
    events_url: str
    labels_url: str
    repository_url: str
    milestone: GHMilestone
    pull_request_links: GHPullRequestLinks
    repository: GHRepository
    reactions: GHReactions
    assignees: List[GHUser]
    node_id: str
    text_matches: List[GHTextMatch]
    active_lock_reason: str

class GHPullRequestComment:
    id: int
    node_id: str
    in_reply_to: int
    body: str
    path: str
    diff_hunk: str
    pull_request_review_id: int
    position: int
    original_position: int
    start_line: int
    line: int
    original_line: int
    original_start_line: int
    side: str
    start_side: str
    commit_id: str
    original_commit_id: str
    user: GHUser
    reactions: GHReactions
    created_at: Timestamp
    updated_at: Timestamp
    author_association: str
    url: str
    html_url: str
    pull_request_url: str
    subject_type: str

class GHReaction:
    id: int
    user: GHUser
    node_id: str
    content: str

class GHRepositoryContentResponse:
    content: GHRepositoryContent
    commit: GHCommit

class GHRepositoryContent:
    type: str
    target: str
    encoding: str
    size: int
    name: str
    path: str
    content: str
    sha: str
    url: str
    git_url: str
    html_url: str
    download_url: str
    submodule_git_url: str


# GHPullRequestAutoMerge represents the "auto_merge" response for a PullRequest.
class GHPullRequestAutoMerge:
    enabled_by: GHUser
    merge_method: str
    commit_title: str
    commit_message: str


# FIXME:
class GHSecurityAndAnalysis:
    pass
class GHLicence:
    pass
class GHTeam:
    pass
class GHPullRequestBranch:
    pass
class GHPRLinks:
    pass
class GHSignatureVerification:
    pass
class GHCommitStats:
    pass
class GHTree:
    pass
class GHCommitAuthor:
    pass
class GHReactions:
    pass
class GHPullRequestLinks:
    pass
class RepositoryContent:
    pass

####################################################################################################


## Issues
def create_issue(owner: str, repo: str, title: str, body: str, assignee: str, milestone: str, labels: str, assignees: str) -> GHIssue:
    """Create an issue
    API:
      see https://docs.github.com/en/rest/issues/issues#create-an-issue
    """
    pass


def get_issue(owner: str, repo: str, number: str) -> GHIssue:
    """Get an issue
    API:
      see https://docs.github.com/en/rest/issues/issues#get-an-issue
    """
    pass

def update_issue(owner: str, repo: str, number: str, title: str, body: str, assignee: str, state: str, stateReason: str, milestone: str, labels: str, assignees: str) -> GHIssue:
    """Update an issue
    API:
      see https://docs.github.com/en/rest/issues/issues#update-an-issue
    """
    pass

def list_repository_issues(owner: str, repo: str, milestone: str, state: str, assignee: str, creator: str, mentioned: str, labels: str, sort: str, direction: str, since: str) -> List[GHIssue]:
    """List repository issues
    API:
      see https://docs.github.com/en/rest/issues/issues#list-repository-issues
    """
    pass

def list_collaborators(owner: str, repo: str, affiliation: str|None, permission: str|None) -> List[GHUser]:
    """List repository collaborators
    API:
      see https://docs.github.com/en/rest/collaborators/collaborators#list-repository-collaborators
    """
    pass

def list_commits(owner: str, repo: str, opts: str|None) -> List[GHRepositoryCommit]
    """List commits
    API:
      see https://docs.github.com/en/rest/commits/commits#list-commits
    """
    pass


def list_workflow_runs(owner: str, repo: str, branch: str|None, event: str|None, actor: str|None, status: str|None, created: str|None, head_sha: str|None, exclude_pull_requests: str|None, check_suite_id: str|None) ->GHWorkflowRuns:
    """List workflow runs
    API:
      see https://docs.github.com/en/rest/issues/issues#list-workflow-runs
    """
    pass


## Issue comments
def create_issue_comment(owner: str, repo: str, number: str, body: str) -> GHIssueComment:
    """Create an issue comment
    API:
      see https://docs.github.com/en/rest/issues/comments#create-an-issue-comment
    """
    pass


## Issue labels
def add_issue_labels(owner: str, repo: str, number: str, labels: str) -> List[GHLabel]
    """Add labes to an issue
    API:
      see https://docs.github.com/en/rest/issues/labels#add-labels-to-an-issue
    """
    pass

def remove_issue_label(owner: str, repo: str, number: str, label: str) -> None:
    """Remove a label from an issue
    API:
      see https://docs.github.com/en/rest/issues/labels#remove-a-label-from-an-issue
    """
    pass


## Pull requests.
def get_pull_request(owner: str, repo: str, number: str) -> GHPullRequest:
    """Get a pull request
    API:
      see https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request
    """
    pass

def list_pull_requests(owner: str, repo: str, state: str, head: str, base: str, sort: str, direction: str) -> List[GHPullRequest]:
    """List pull requests
    API:
      see https://docs.github.com/en/rest/pulls/pulls#list-pull-requests
    """
    pass

def create_pull_request(owner: str, repo: str, head: str, base: str, title: str|None, body: str|None, head_repo: str|None, draft: str|None, issue: str|None, maintainer_can_modify: bool|None) -> GHPullRequest:
    """Create a pull request
    API:
      see https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request
    """
    pass

def request_review(owner: str, repo: str, number: str, reviewers: str|None, team_reviewers: str|None) -> GHPullRequest:
    """Request review for a pull request
    API:
      see https://docs.github.com/en/rest/pulls/review-requests#request-reviewers-for-a-pull-request
    """
    pass

## Pull-request comments.
def list_review_comments(owner: str, repo: str, number: str) -> List[GHPullRequestComment]
    """List review comments on a pull request
    API:
      see https://docs.github.com/en/rest/pulls/comments#list-review-comments-on-a-pull-request
    """
    pass

## Reactions.
def create_reaction_for_commit_comment(owner: str, repo: str, id: str, content: str) -> GHReaction:
    """Creates reaction for a commit comment
    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-commit-comment
    """
    pass

def create_reaction_for_issue(owner: str, repo: str, number: str, content: str) -> GHReaction:
    """Creates reaction for an issue
    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue"""
    pass

def create_reaction_for_issue_comment(owner: str, repo: str, id: str, content: str) -> GHReaction:
    """Create reaction for an issue comment
    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-an-issue-comment
    """
    pass

def create_reaction_for_pull_request_review_comment(owner: str, repo: str, id: str, content: str) -> GHReaction:
    """Create reaction for a pull request review comment
    API:
      see https://docs.github.com/en/rest/reactions/reactions#create-reaction-for-a-pull-request-review-comment
    """
    pass


## Repository Contents.
def create_file(owner: str, repo: str, path: str, content: str, message: str, sha: str|None, branch: str|None, committer: str|None) -> GHRepositoryContentResponse:
    """Create or update file contents
    API:
      see https://docs.github.com/en/rest/repos/contents#create-or-update-file-contents
    """
    pass

def get_contents(owner: str, repo: str, path: str, ref: str|None) -> List[GHRepositoryContent]:
    """Get repository content
    API:
      see https://docs.github.com/en/rest/repos/contents#get-repository-content
    """
    pass

## Git references.
def create_ref(owner: str, repo: str, ref: str, sha: str) -> GHReference:
    """Create a reference
    API:
      see https://docs.github.com/en/rest/git/refs#create-a-reference
    """
    pass

def get_ref(owner: str, repo: str, ref: str) ->GHReference:
    """Get a reference
    API:
      see https://docs.github.com/en/rest/git/refs#get-a-reference
    """
    pass


## Users
def get_user(username: str) -> GHUser:
    """Get a user
    API:
      see https://docs.github.com/en/rest/users#get-a-user
    """
    pass
