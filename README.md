# Project Rare Life - MVP Specification

## Pages

1. Home
Mission statement and entry point to form

2. Survey
Add blocks for describing symptoms

3. Account profile page to view history form intake form
Add new symptoms or diagnoses and close out or modify previous info

### Survey

#### Account fields

- Email
- Generate fun / unique username

#### User flow

1. Add symptoms
	a. Add treatment for symptom
        - Outcomes of treatment?
	b. Add change in symptom

#### Symptom fields

- Symptom name: Enum
- Age of appearance
	- Years: Int
	- Months: Int
- Symptom severity: Int

#### Symptom categories

- Medical (e.g. microcephely)
- Behavioral (e.g. hand flapping)

#### Treatment fields

- Treatment name: Enum
- Age of application
	- Years: Int
	- Months: Int

## Structure

Event sources architecture to allow users to submit events around symptoms

### Events

- Symtpom noted
- Symptom improved
- Symptom worsed
- Symptom resolved
