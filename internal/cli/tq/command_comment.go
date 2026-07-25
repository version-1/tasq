package tq

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (a app) routeComment(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printCommentHelp(a.stdout)
		return nil
	}
	action := args[0]
	switch action {
	case "add":
		return a.commentAdd(ctx, args[1:], cfg)
	case "list":
		return a.commentList(ctx, args[1:], cfg)
	default:
		return usageError("unknown comment action %q", action)
	}
}

func (a app) commentAdd(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 {
		return usageError("usage: tq comment add <issue-id> [flags]")
	}
	issueID, err := parseID(args[0])
	if err != nil {
		return err
	}
	fs := newFlagSet("comment add")
	author := fs.String("author", defaultCommentAuthor(), "comment author")
	commentType := fs.String("type", string(entity.CommentGeneral), "comment type")
	body := fs.String("body", "", "comment body")
	attach := fs.String("attach", "", "image attachment path")
	if err := fs.Parse(args[1:]); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("comment add does not accept extra positional arguments")
	}
	if *author == "" {
		return usageError("author is required")
	}
	if *body == "" {
		return usageError("body is required")
	}
	input := entity.CreateCommentInput{
		Author: *author,
		Type:   entity.CommentType(*commentType),
		Body:   *body,
	}
	comment, err := a.client.createComment(ctx, issueID, input)
	if err != nil {
		return err
	}
	if *attach != "" {
		attachment, err := a.client.uploadAttachment(ctx, attachmentUploadInput{
			EntityType: entity.AttachmentEntityComment,
			EntityID:   strconv.FormatInt(comment.ID, 10),
			Path:       *attach,
		})
		if err != nil {
			return err
		}
		body := appendAttachmentMarkdown(comment.Body, attachment)
		comment, err = a.client.updateComment(ctx, comment.ID, entity.UpdateCommentInput{Body: &body})
		if err != nil {
			_ = a.client.deleteAttachment(ctx, attachment.ID)
			return err
		}
	}
	return writeComment(a.stdout, cfg.output, comment)
}

func (a app) commentList(ctx context.Context, args []string, cfg config) error {
	if len(args) != 1 {
		return usageError("usage: tq comment list <issue-id>")
	}
	issueID, err := parseID(args[0])
	if err != nil {
		return err
	}
	comments, err := a.client.listComments(ctx, issueID)
	if err != nil {
		return err
	}
	return writeComments(a.stdout, cfg.output, comments)
}

func appendAttachmentMarkdown(content string, attachment entity.Attachment) string {
	markdown := fmt.Sprintf("![%s](attachment://%s)", markdownAltText(attachment.Filename), attachment.ID)
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return markdown
	}
	return content + "\n\n" + markdown
}

func markdownAltText(value string) string {
	value = strings.ReplaceAll(value, "[", "")
	value = strings.ReplaceAll(value, "]", "")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}
