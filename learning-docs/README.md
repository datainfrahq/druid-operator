# Apache Druid Kubernetes Operator - Complete Learning Guide

Welcome! This documentation is designed for developers who are new to Kubernetes operators, Go programming, and want to understand how this Apache Druid Operator works from the ground up.

## Table of Contents

1. [Introduction](./01-introduction/README.md) - What is this project?
2. [Prerequisites & Learning Path](./02-prerequisites/README.md) - What you need to learn
3. [Go vs Java](./03-go-vs-java/README.md) - Understanding Go for Java developers
4. [Kubernetes Fundamentals](./04-kubernetes-fundamentals/README.md) - K8s concepts you need
5. [What is an Operator?](./05-operators/README.md) - The operator pattern explained
6. [Custom Resources (CR)](./06-custom-resources/README.md) - Understanding CRDs and CRs
7. [Project Structure](./07-project-structure/README.md) - How this codebase is organized
8. [The Reconciliation Loop](./08-reconciliation-loop/README.md) - Core operator logic
9. [Apache Druid Overview](./09-druid-overview/README.md) - What is Druid?
10. [Complete Flow Walkthrough](./10-complete-flow/README.md) - End-to-end explanation

**Bonus:** [Quick Reference Guide](./quick-reference.md) - Commands and cheat sheet

---

## Quick Start

If you're completely new, read the documents in order. Each builds on the previous one.

If you have some experience:
- Java developer new to Go? Start with [Go vs Java](./03-go-vs-java/README.md)
- Know Go but new to K8s? Start with [Kubernetes Fundamentals](./04-kubernetes-fundamentals/README.md)
- Know K8s but new to operators? Start with [What is an Operator?](./05-operators/README.md)

---

## What This Repository Does

This is a **Kubernetes Operator** that automates the deployment and management of **Apache Druid** clusters on Kubernetes.

In simple terms:
- **Apache Druid** = A database for real-time analytics
- **Kubernetes** = A platform to run containerized applications
- **Operator** = A program that automates complex application management on Kubernetes

Instead of manually creating dozens of Kubernetes resources (Pods, Services, ConfigMaps, etc.) to run Druid, you just write ONE simple YAML file, and this operator creates everything for you automatically!

