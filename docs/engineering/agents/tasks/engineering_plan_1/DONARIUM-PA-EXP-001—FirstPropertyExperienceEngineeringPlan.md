# DONARIUM-PA-EXP-001 — First Property Experience Engineering Plan

## Role

Operate as the OSK Product Architect for Donarium.

Use the Product Architect role contract and the capabilities available in the current OSK workspace.

## Objective

Produce an Engineering Plan Proposal for the Donarium **First Property Experience**.

This is a planning task, not an implementation task.

Do not write production code or begin implementation.

## Evidence available

Inspect and reconcile the relevant evidence available in the workspace, including:

1. The current Donarium repository and implementation.
2. Existing Donarium documentation, roadmap, ADRs, migrations, and engineering artifacts.
3. The OSK workspace, including applicable Roles and Skills.
4. The Product Design Brief for the First Property Experience, if present in the workspace.
5. The current Donarium product design available through the connected Stitch MCP server.

Do not assume that any single source represents the complete or canonical system state.

## Stitch

A Stitch MCP server is connected to this Claude Code project.

Use it to inspect the Donarium design relevant to the First Property Experience.

Treat Stitch as product/design evidence, not as unquestionable specification.

Distinguish, where relevant, between:

- explicit product decisions;
- design hypotheses introduced by Stitch;
- behavior already supported by the implementation;
- existing architectural/domain constraints;
- unresolved product questions.

Do not silently convert a Stitch design choice into a product or engineering requirement.

## Scope

The target experience is the **First Property Experience**, including the relevant paths around:

- zero accessible properties;
- adding a first property;
- accessing an existing property;
- establishing ownership and management relationships;
- the one-property experience;
- the multi-property Portfolio;
- Property Home;
- property-level information organization;
- payment state and payment registration;
- property-specific events/attention;
- direct/deep-link entry where relevant.

Determine the actual engineering boundary from the evidence rather than assuming every visible Stitch feature belongs in the first implementation slice.

## Planning expectations

Your proposal should:

- establish the current system baseline before proposing changes;
- identify existing capabilities that can be reused;
- identify missing domain/application/UI capabilities;
- reconcile product design with the existing implementation;
- surface meaningful contradictions or gaps;
- identify domain concepts only when supported by evidence;
- avoid speculative infrastructure or premature abstractions;
- preserve existing architectural boundaries unless evidence justifies changing them;
- decompose the work into meaningful, preferably vertical, work slices;
- identify dependencies between work slices;
- identify work that can safely proceed in parallel;
- assign appropriate OSK Roles to work, rather than directly assigning Skills;
- identify verification/review needs;
- preserve intentionally unresolved product questions rather than inventing policy.

Ask questions only when an unresolved issue materially prevents a useful engineering plan.

If a reasonable assumption allows planning to continue, make the assumption explicit and classify it appropriately instead of blocking unnecessarily.

## Expected output

Produce an **Engineering Plan Proposal** suitable for review by the Product Authority.

The proposal should make clear:

1. What exists today.
2. What product experience is being introduced.
3. What important evidence was found in Stitch.
4. Where repository reality and product/design evidence agree or disagree.
5. What domain/application/UI changes appear necessary.
6. The proposed implementation slices.
7. Dependencies and parallelization opportunities.
8. Which OSK Role should own each slice or review activity.
9. Material risks, ambiguities, assumptions, and open product decisions.
10. What should explicitly remain out of scope.

Include enough reasoning that another engineer can understand why the plan has this shape.

## Stop condition

STOP after producing the Engineering Plan Proposal.

Do not:

- implement the plan;
- modify production code;
- create migrations;
- refactor the system;
- resolve open product questions by fiat;
- dispatch implementation work;
- invoke implementation agents.

The proposal requires Product Authority review and approval before execution.
