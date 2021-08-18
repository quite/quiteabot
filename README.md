
This is a tiny two-way Telegram-bot to XMPP-bridge. I has worked well for me
for talking to some people that prefer Telegram.

# Setup

To set it up, you need to:

- `cp quiteabot.yaml.example quiteabot.yaml`
- Fill out the bot's XMPP account (`xmppuser`, `xmpppass`)
  - The XMPP server host & port is found by looking up the DNS SRV record
    associated with the host part of the `xmppuser` (jabber ID). `xmppserver`
    can be set to override this.
- Set `xmpptarget` to your own jid, or so
- Get a Telegram account (mobile phone number needed)
- Register a Telegram bot: https://core.telegram.org/bots
- Fill in authorization token (`telegramtoken`)
- Map some Telegram userids to names in the config (actually conversation ids,
  but I don't think they change) (`telegramusers`)

# Usage

Your friend needs to:

- Initially chat up your Telegram bot by its name and tell it `/start`
- Message the bot to message you

*quiteabot* will forward the message to the configured `xmpptarget`, resolving
the name from the userid, if present in `telegramusers`.

To message a Telegram user, you need to:

- Send a message over XMPP to your bot, prefixing it with a `username:` from
  the mapping

# TODO

Can only message a user by name, so first need to stuff them into `telegramusers`.
