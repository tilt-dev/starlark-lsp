## Auth
def auth_test():
    """https://api.slack.com/methods/auth.test"""
    pass


## Chat
def chat_delete(channel: str, ts: str):
    """https://api.slack.com/methods/chat.delete"""
    pass

def chat_post_ephemeral(channel: str, user: str, text: str, blocks: str|None, thread_ts: str|None):
    """https://api.slack.com/methods/chat.postEphemeral"""
    pass

def chat_post_message(channel: str, text: str|None, blocks: str|None, thread_ts: str|None, reply_broadcast: str|None):
    """https://api.slack.com/methods/chat.postMessage"""
    pass

def chat_update(channel: str, ts: str, text: str|None, blocks: str|None, reply_broadcast: str|None):
    """https://api.slack.com/methods/chat.update"""
    pass

def send_text_message(target: str, text: str, thread_ts: str|None, reply_broadcast: str|None):
    """convenience wrapper for chat.postMessage"""
    pass

def send_approval_message(target: str, header: str, message: str, green_button: str|None, red_button: str|None, thread_ts: str|None, reply_broadcast: str|None):
    """convenience wrapper for chat.postMessage"""
    pass


## Conversations
def conversations_history(channel: str, cursor: str|None, limit: str|None, include_all_metadata: str|None, inclusive: str|None, oldest: str|None, latest: str|None):
    """https://api.slack.com/methods/conversations.history"""
    pass

def conversations_info(channel: str, include_locale: str|None, include_num_members: str|None):
    """https://api.slack.com/methods/conversations.info"""
    pass

def conversations_list(cursor: str|None, limit: str|None, exclude_archived: str|None, team_id: str|None, types: str|None):
    """https://api.slack.com/methods/conversations.list"""
    pass

def conversations_replies(channel: str, ts: str, cursor: str|None, limit: str|None, include_all_metadata: str|None, inclusive: str|None, oldest: str|None, latest: str|None):
    """https://api.slack.com/methods/conversations.replies"""
    pass


## Reactions.
def reactions_add(channel: str, name: str, timestamp: str):
    """https://api.slack.com/methods/reactions.add"""
    pass


## Users.
def users_get_presence(user: str|None):
    """https://api.slack.com/methods/users.getPresence"""
    pass

def users_info(user: str, include_locale: str|None):
    """https://api.slack.com/methods/users.info"""
    pass

def users_list(cursor: str|None, limit: str|None, include_locale: str|None, team_id: str|None):
    """https://api.slack.com/methods/users.list"""
    pass

def users_lookup_by_email(email: str):
    """https://api.slack.com/methods/users.lookupByEmail"""
    pass
