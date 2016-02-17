# GQL
A GraphQL library for GoLang. Currently in alpha.

[![GoDoc](https://godoc.org/github.com/jmank88/gql?status.svg)](https://godoc.org/github.com/jmank88/gql) [![Build Status](https://travis-ci.org/jmank88/gql.svg)](https://travis-ci.org/jmank88/gql)

Born out of a port of the reference javascript implementation: https://github.com/graphql/graphql-js

This implementation is intended to maximize usability and efficiency.

- package lang
  - [x] package ast

  - [x] package parser
    - [x] tests
    - [ ] benchmarks

    - [x] package lexer
        - [x] tests
        - [x] benchmarks

        - [x] package scanner
          - [x] tests
          - [x] benchmarks

    - [x] package token

  - [x] package printer
    - [ ] tests

- [ ] package validation

- [ ] package execution

- [ ] package server

- [ ] package client

- [ ] package encoding
  - [ ] json
  - [ ] grpc
  - [ ] Cap'n proto

- package cmd
  - [ ] gql fmt