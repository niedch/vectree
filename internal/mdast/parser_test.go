package mdast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMarkdown_EmptyInput(t *testing.T) {
	markdown := ""
	doc := ParseMarkdown(markdown)

	require.NotNil(t, doc, "Expected non-nil document")

	expected := `Document
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_SingleHeading(t *testing.T) {
	markdown := `# Hello World`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Hello World"
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_MultipleHeadingLevels(t *testing.T) {
	markdown := `# Title
## Subtitle
### Section`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Title"
  Heading (level=2)
    Text: "Subtitle"
  Heading (level=3)
    Text: "Section"
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_SingleParagraph(t *testing.T) {
	markdown := `This is a paragraph.`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Paragraph
    Text: "This is a paragraph."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_MultiLineParagraph(t *testing.T) {
	markdown := `Line one.
Line two.
Line three.`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Paragraph
    Text: "Line one."
    Text: "Line two."
    Text: "Line three."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_HeadingAndParagraph(t *testing.T) {
	markdown := `# Introduction

This is the introduction paragraph.`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Introduction"
  Paragraph
    Text: "This is the introduction paragraph."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_ComplexDocument(t *testing.T) {
	markdown := `# Main Title

First paragraph line one.
First paragraph line two.

## Section One

Section one content.

## Section Two

Section two line one.
Section two line two.`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Main Title"
  Paragraph
    Text: "First paragraph line one."
    Text: "First paragraph line two."
  Heading (level=2)
    Text: "Section One"
  Paragraph
    Text: "Section one content."
  Heading (level=2)
    Text: "Section Two"
  Paragraph
    Text: "Section two line one."
    Text: "Section two line two."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_BlankLinesIgnored(t *testing.T) {
	markdown := `# Title

Paragraph text.


`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Title"
  Paragraph
    Text: "Paragraph text."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_AllHeadingLevels(t *testing.T) {
	markdown := `# Level 1
## Level 2
### Level 3
#### Level 4
##### Level 5
###### Level 6`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Level 1"
  Heading (level=2)
    Text: "Level 2"
  Heading (level=3)
    Text: "Level 3"
  Heading (level=4)
    Text: "Level 4"
  Heading (level=5)
    Text: "Level 5"
  Heading (level=6)
    Text: "Level 6"
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_RealWorldExample(t *testing.T) {
	markdown := `# Getting Started

Welcome to our documentation. This guide will help you get started quickly.

## Installation

To install the software, run the following command:

### Prerequisites

Make sure you have the following installed:

- Go 1.20 or higher
- Git

### Download

Clone the repository and build the project.

## Configuration

After installation, you'll need to configure the application.

Create a configuration file in your home directory.

## Usage

Now you're ready to use the application!

### Basic Commands

Start with these basic commands to familiarize yourself.

### Advanced Features

Once comfortable, explore these advanced features.

# Troubleshooting

If you encounter issues, check this section.

## Common Problems

Here are solutions to common problems.

## Getting Help

Contact support if you need additional assistance.`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "Getting Started"
  Paragraph
    Text: "Welcome to our documentation. This guide will help you get started quickly."
  Heading (level=2)
    Text: "Installation"
  Paragraph
    Text: "To install the software, run the following command:"
  Heading (level=3)
    Text: "Prerequisites"
  Paragraph
    Text: "Make sure you have the following installed:"
  Paragraph
    Text: "- Go 1.20 or higher"
    Text: "- Git"
  Heading (level=3)
    Text: "Download"
  Paragraph
    Text: "Clone the repository and build the project."
  Heading (level=2)
    Text: "Configuration"
  Paragraph
    Text: "After installation, you'll need to configure the application."
  Paragraph
    Text: "Create a configuration file in your home directory."
  Heading (level=2)
    Text: "Usage"
  Paragraph
    Text: "Now you're ready to use the application!"
  Heading (level=3)
    Text: "Basic Commands"
  Paragraph
    Text: "Start with these basic commands to familiarize yourself."
  Heading (level=3)
    Text: "Advanced Features"
  Paragraph
    Text: "Once comfortable, explore these advanced features."
  Heading (level=1)
    Text: "Troubleshooting"
  Paragraph
    Text: "If you encounter issues, check this section."
  Heading (level=2)
    Text: "Common Problems"
  Paragraph
    Text: "Here are solutions to common problems."
  Heading (level=2)
    Text: "Getting Help"
  Paragraph
    Text: "Contact support if you need additional assistance."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}

func TestParseMarkdown_MixedContent(t *testing.T) {
	markdown := `# API Documentation

The API provides RESTful endpoints for managing resources.

## Authentication

All requests require authentication.
Use Bearer tokens in the Authorization header.

### Getting a Token

Request a token from the /auth endpoint.
Include your credentials in the request body.

## Endpoints

### Users

#### GET /users

Returns a list of all users.

#### POST /users

Creates a new user.

### Products

#### GET /products

Returns all products.
Supports pagination and filtering.`

	doc := ParseMarkdown(markdown)

	expected := `Document
  Heading (level=1)
    Text: "API Documentation"
  Paragraph
    Text: "The API provides RESTful endpoints for managing resources."
  Heading (level=2)
    Text: "Authentication"
  Paragraph
    Text: "All requests require authentication."
    Text: "Use Bearer tokens in the Authorization header."
  Heading (level=3)
    Text: "Getting a Token"
  Paragraph
    Text: "Request a token from the /auth endpoint."
    Text: "Include your credentials in the request body."
  Heading (level=2)
    Text: "Endpoints"
  Heading (level=3)
    Text: "Users"
  Heading (level=4)
    Text: "GET /users"
  Paragraph
    Text: "Returns a list of all users."
  Heading (level=4)
    Text: "POST /users"
  Paragraph
    Text: "Creates a new user."
  Heading (level=3)
    Text: "Products"
  Heading (level=4)
    Text: "GET /products"
  Paragraph
    Text: "Returns all products."
    Text: "Supports pagination and filtering."
`
	actual := PrintAST(doc, 0)
	assert.Equal(t, expected, actual, "AST mismatch")
}
