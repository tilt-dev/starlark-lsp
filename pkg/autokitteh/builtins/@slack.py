####################################################################################################
## Classes
class SlackMessage:
    type: str
    subtype: str
    hidden: bool

    text: str
    blocks: List[SlackBlock]
    edited: SlackEdited|None

    user: str
    app_id: str
    bot_id: str
    bot_profile: BotProfile|None
    parent_user_id: str

    team: str
    channel: str
    channel_type: str
    ts: str
    event_ts: str
    permalink: str

    reply_count: int
    reply_users_count: int
    latest_reply: str
    reply_users: List[str]
    last_read: str
    unread_count: int

    unread_count_display: int
    is_locked: bool
    subscribed: bool

    is_starred: bool
    pinned_to: List[str]
    reactions: List[SlackReaction]

    inviter: str
    name: str
    old_name: str
    purpose: str
    topic: str

    message: SlackMessage|None
    previous_message: SlackMessage|None
    deleted_ts: str
    thread_ts: str

    root: SlackMessage|None

    client_msg_id: str

class SlackReaction:
    name: str
    users: List[str]
    count: int

class SlackText:
    type: str
    text: str
    emoji: bool
    verbatim: bool
# FIXME: should we allow to pass SlackText and not its serialized version, i.e. text:str

class SlackChannel:
    id: str
    name: str
    name_normalized: str
    previous_names: List[str]
    creator: str
    user: str

    is_member: bool
    is_read_only: bool
    is_thread_only: bool

    topic: SlackTopic|None
    purpose: SlackPurpose|None
    last_read: str
    unread_count: int
    unread_count_display: int

    is_archived: bool
    is_channel: bool
    is_frozen: bool
    is_general: bool
    is_group: bool
    is_im: bool
    is_mpim: bool
    is_open: bool
    is_private: bool

    is_shared: bool
    is_org_shared: bool
    is_ext_shared: bool
    is_pending_ext_shared: bool

    context_team_id: str
    shared_team_ids: List[str]
    pending_connected_team_ids: List[str]

    created: int
    updated: int
    unlinked: int

    latest: SlackMessage|None

    locale: str
    num_members: int
    priority: float

class SlackConversationError:
    user: str
    ok: bool
    error: str

class SlackConversationPurposeOrTopic:
    value: str
    creator: str
    last_set: int
# FIXME: could we pass this as param to set_topic and and set_purpose


class SlackEnterpriseUser:
    enterprise_id: str
    enterprise_name: str
    id: str
    is_admin: bool
    is_owner: bool
    teams: List[str]

class SlackProfile:
    first_name: str
    last_name: str
    real_name: str
    real_name_normalized: str
    display_name: str
    display_name_normalized: str

    email: str
    phone: str
    title: str
    pronouns: str
    start_date: str
    team: str

    api_app_id: str
    bot_id: str
    always_active: bool

    status_text: str
    status_text_canonical: str
    status_emoji: str
    status_expiration: int

    is_custom_image: bool
    image_original: str
    image_24: str
    image_32: str
    image_48: str
    image_72: str
    image_192: str
    image_512: str
    image_1024: str
    avatar_hash: str

class SlackUser:
    id: str
    team_id: str
    real_name: str

    profile: SlackProfile|None
    enterprise_user: SlackEnterpriseUser|None

    deleted: bool
    is_admin: bool
    is_app_user: bool
    is_bot: bool
    is_email_confirmed: bool
    is_invited_user: bool
    is_owner: bool
    is_primary_owner: bool
    is_restricted: bool
    is_stranger: bool
    is_ultra_restricted: bool

    tz: str
    tz_label: str
    tz_offset: int

    updated: int


####################################################################################################
## Responses
class SlackResponse:
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict

class SlackAuthTest:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    url: str
    team: str
    user: str
    team_id: str
    user_id: str
    bot_id: str|None
    enterprise_id:  str|None
    is_enterprise_install: bool

class SlackChatDelete:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: str
    ts: str

class SlackPostEphemeral:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: str
    message_ts: str
# FIXME: message_ts or ts?
# FIXME: I don't see that Slack API returns channel

class SlackPostMessage:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: str
    ts: str
    message: SlackMessage

class SlackUpdate:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: str
    ts: str
    text: str
    message: dict
# FIXME: message struct

class SlackSendTextMessage:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: str
    ts: str
    message: SlackMessage

class SlackSendApprovalMessage:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: str
    ts: str
    message: SlackMessage


class SlackConversationsCreate:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: SlackChannel|None

class SlackConversationsHistory:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    messages: List[SlackMessage]
    has_mode: bool
    pin_count: int
    channel_actions_count: int

class SlackConversationsInfo:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: SlackChannel|None

class SlackConversationsInvite:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: SlackChannel|None
    errors: List[SlackConversationError]

class SlackConversationsList:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channels: List[SlackChannel]

class SlackConversationsMembers:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    members: List[str]


class SlackConversationsOpen:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: SlackChannel|None
    errors: List[SlackConversationError]

class SlackConversationsRename:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    channel: SlackChannel|None

class SlackConversationsReplies:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    messages: List[SlackMessage]
    has_more: bool

class SlackUsersGetPresence:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    online: bool
    auto_away: bool
    manual_away: bool
    connection_count: int

class SlackUsersInfo:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    user: SlackUser|None

class SlackUsersList:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    offset: string
    cache_ts: int
    members: List[SlackUser]


class SlackUsersLookupByEmail:
    # SlackResponse
    ok: bool
    error: str|None
    warning: str|None
    responce_metadata: dict
    # ------------------------
    user: SlackUser|None

####################################################################################################
## Auth
def auth_test() -> SlackAuthTest:
    """Checks authentication and tells "you" who you are, even if you might be a bot.
       [API](https://api.slack.com/methods/auth.test)

    Returns:
      SlackAuthTest
    """
    pass


## Chat
def chat_delete(channel: str, ts: str) -> SlackChatDelete:
    """Deletes a message from a conversation.
       When used with a user token, this method may only delete messages that user themselves can delete in Slack.
       When used with a bot token, this method may delete only messages posted by that bot.
       The response includes the `channel` and `timestamp` properties of the deleted message.
       [API](https://api.slack.com/methods/chat.delete)

    Returns:
      SlackChatDelete
    """
    pass

def chat_post_ephemeral(channel: str, user: str, text: str, blocks: str|None, thread_ts: str|None) -> SlackPostEphemeral:
    """Posts an ephemeral message, which is visible only to the assigned user in a specific public channel, private channel, or private conversation.
    Ephemeral message delivery is not guaranteed — the user must be currently active in Slack and a member of the specified `channel`. By nature, ephemeral messages do not persist across reloads, desktop and mobile apps, or sessions. Once the session is closed, ephemeral messages will disappear and cannot be recovered.
    Use ephemeral messages to send users context-sensitive messages, relevant to the channel they're detectably participating in. Avoid sending unexpected or unsolicited ephemeral messages.

    text or blocks
    The usage of the `text` field changes depending on whether you're using `blocks`. If you're using `blocks`, this is used as a fallback string to display in notifications. If you aren't, this is the main body text of the message. It can be formatted as plain text, or with `mrkdwn`.
    The `text` field is not enforced as required when using `blocks`. However, we highly recommended that you include `text` to provide a fallback when using `blocks`, as described above.

    Formatting
    Messages are formatted as described in the [formatting spec](https://api.slack.com/docs/message-formatting).
    For best results, limit the number of characters in the `text` field to a few thousand bytes at most. Ideally, messages should be short and human-readable, if you need to post longer messages, please consider [uploading a snippet instead](https://api.slack.com/methods/files.upload). (A single message should be no larger than 4,000 bytes.)
    Consider reviewing our [message guidelines](https://api.slack.com/docs/message-guidelines).

    Authorship
    How message authorship is attributed varies by a few factors, with some behaviors varying depending on the kinds of tokens you're using to post a message.

    .... Many more ...

    API
    [API](https://api.slack.com/methods/chat.postEphemeral)

    Returns:
      SlackChatPostEphemeral
    """
    # FIXME: description is too long?
    # FIXME: how to describe args better, e.g. text or blocks
    pass

def chat_post_message(channel: str, text: str|None, blocks: str|None, thread_ts: str|None, reply_broadcast: str|None) -> SlackPostMessage:
    """https://api.slack.com/methods/chat.postMessage

    Returns:
      SlackPostMessage
    """
    pass

def chat_update(channel: str, ts: str, text: str|None, blocks: str|None, reply_broadcast: str|None) -> SlackChatUpdate:
    """https://api.slack.com/methods/chat.update

    Returns:
      SlackUpdate
    """
    pass

def send_text_message(target: str, text: str, thread_ts: str|None, reply_broadcast: str|None) -> SlackSendTextMessage:
    """convenience wrapper for chat.postMessage

    Returns:
      SlackSendTextMessage
    """
    pass

def send_approval_message(target: str, header: str, message: str, green_button: str|None, red_button: str|None, thread_ts: str|None, reply_broadcast: str|None) -> SlackSendApprovalMessage:
    """convenience wrapper for chat.postMessage

    Returns:
      SlackSendApprovalMessage
    """
    pass


## Conversations
def conversations_archive(channel: str) -> SlackResponse:
    """https://api.slack.com/methods/conversations.archive

    Returns:
      SlackResponse
    """
    pass

def conversations_close(channel: str) -> SlackResponse:
    """https://api.slack.com/methods/conversations.close

    Returns:
      SlackResponse
    """
    pass

def conversations_create(name: str, is_private: bool|None, team_id: str|None) -> SlackConversationsCreate:
    """https://api.slack.com/methods/conversations.create

    Returns:
      SlackConversationsCreate
    """
    pass

def conversations_history(channel: str, cursor: str|None, limit: str|None, include_all_metadata: bool|None, inclusive: bool|None, oldest: str|None, latest: str|None) -> SlackConversationsHistory:
    """https://api.slack.com/methods/conversations.history

    Returns:
      SlackConversationsHistory
    """
    pass

def conversations_info(channel: str, include_locale: bool|None, include_num_members: bool|None) -> SlackConversationsInfo:
    """https://api.slack.com/methods/conversations.info

    Returns:
      SlackConversationsInfo
    """
    pass

def conversations_invite(channel: str, users: str, force: bool|None) -> SlackConversationsInvite:
    """https://api.slack.com/methods/conversations.invite

    Returns:
      SlackConversationsInvite
    """
    pass

def conversations_list(cursor: str|None, limit: int|None, exclude_archived: bool|None, team_id: str|None, types: str|None) -> SlackConversationsList:
    """https://api.slack.com/methods/conversations.list

    Returns:
      SlackConversationsList
    """
    pass

def conversations_members(channel:str, cursor: str|None, limit: int|None) -> SlackConversationsMembers:
    """https://api.slack.com/methods/conversations.members

    Returns:
      SlackConversationsMembers
    """
    pass

def conversations_open(channel:str|None, users: str|None. prevent_creation: bool|None) -> SlackConversationsOpen:
    """https://api.slack.com/methods/conversations.open

    Returns:
      SlackConversationsOpen
    """
    pass

def conversations_rename(channel:str, name: str) -> SlackConversationsRename:
    """https://api.slack.com/methods/conversations.rename

    Returns:
      SlackConversationsRename
    """
    pass

def conversations_replies(channel: str, ts: str, cursor: str|None, limit: int|None, include_all_metadata: bool|None, inclusive: bool|None, oldest: str|None, latest: str|None) -> SlackConversationsReplies:
    """https://api.slack.com/methods/conversations.replies

    Returns:
      SlackConversationsReplies
    """
    pass

def conversations_set_purpose(channel:str, purpose: str) -> SlackResponse:
    """https://api.slack.com/methods/conversations.setPurpose

    Returns:
      SlackResponse
    """
    pass

def conversations_set_topic(channel:str, topic: str) -> SlackResponse:
    """https://api.slack.com/methods/conversations.setTopic

    Returns:
      SlackResponse
    """
    pass

def conversations_unarchive(channel:str) -> SlackResponse:
    """https://api.slack.com/methods/conversations.unarchive

    Returns:
      SlackResponse
    """
    pass


## Reactions.
def reactions_add(channel: str, name: str, timestamp: str) -> SlackResponse:
    """https://api.slack.com/methods/reactions.add

    Returns:
      SlackResponse
    """
    pass


## Users.
def users_get_presence(user: str|None) -> SlackUsersGetPresence:
    """https://api.slack.com/methods/users.getPresence

    Returns:
      SlackUsersGetPresence
    """
    pass

def users_info(user: str, include_locale: bool|None) -> SlackUsersInfo:
    """https://api.slack.com/methods/users.info

    Returns:
      SlackUsersInfo
    """
    pass

def users_list(cursor: str|None, limit: int|None, include_locale: bool|None, team_id: str|None) -> SlackUsersList:
    """https://api.slack.com/methods/users.list

    Returns:
      SlackUsersList
    """
    pass

def users_lookup_by_email(email: str) -> SlackUsersLookupByEmail:
    """https://api.slack.com/methods/users.lookupByEmail

    Returns:
      SlackUsersLookupByEmail
    """
    pass
