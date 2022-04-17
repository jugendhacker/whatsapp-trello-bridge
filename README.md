# WhatsApp-Trello Bridge

A small program, for the connection of WhatsApp and Trello

## Modes
The code has two modes:
1) ticket based

    *When a message is written to the WhatsApp user, a new ticket (in the form of a new card) is opened, to which other messages are attached. This ticket can be closed by moving it to a specific list, and a new ticket is created for further WhatsApp messages.*

    Branch: `tickets`
2) user based

    *A card is created in the Trello board for each WhatsApp user. When new messages are received, this card is always moved to the "Unread" list.*

    Branch: `bridge`

## Setup

Clone the repo and switch to the branch of the mode you want to use.

Rename `example.env` to `.env`, fill out (or modify) the environment variables and run the tool with `go run main.go`
