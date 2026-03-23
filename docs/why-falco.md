# Why Falco?

> Terraform tells us **what should exist**.
> Falco tells us **what actually happened**.

---

## The Perfect Blueprint and The Witness

Imagine a city with a brilliant architect who meticulously documents everything in a **blueprint** (Terraform). Every building, every road, every gate — perfectly mapped out. The blueprint represents the "ideal city."

But one night, someone secretly replaces a gate. By morning, a lock has been added. Yet the blueprint shows no changes.

The architect walks through the city the next day, comparing reality to the blueprint. Finally, they notice: *"...ah, it's different."*

But it's too late. They can see:

- **What** changed

But not:

- **Who** did it
- **When** it happened
- **Why** it was done

**The blueprint only speaks of results, not actions.**

---

## Enter Falco: The Witness

So the city hires **Falco** — not an architect, not a designer, but a **witness**.

Falco's job is singular and essential:

> **To observe the exact moment someone takes action.**

Falco doesn't build. Falco doesn't draw maps. Falco watches:

- **Who** touched the gate
- **When** they did it
- **Which** gate it was
- **What** their intent was

Not after the change — **during the moment of change**.

One midnight, Falco observes an unfamiliar person approaching the gate from an unusual path, reaching for the lock. At that instant, Falco alerts:

> *"Now. This is not the city we know."*

---

## The Meeting of Blueprint and Witness

The architect listens to Falco's report:

- *"Who touched it?"*
- *"When?"*
- *"Which gate?"*

The architect opens the blueprint and realizes:

> **"That change... doesn't exist in my blueprint."**

In that moment, they both understand:

| Role | What it knows |
|------|---------------|
| **The Blueprint** (Terraform) | *What should exist* |
| **The Witness** (Falco) | *What actually happened* |

Neither alone can protect the city.

---

## What Makes Falco Special

Falco speaks of **actions, not just results**:

- Not states, but **behaviors**
- Not diffs, but **intentions**
- Not resources, but **people**

That's why Falco can answer:

> *"Not just 'what happened in this city,' but 'why it happened.'"*

---

## The Modern City (Cloud)

Today's cities aren't built by humans alone:

- Bots and CI/CD pipelines
- Automation scripts
- AI agents

Changes happen in an instant. That's why we need more than post-mortem audits. We need **someone who was there when it happened**.

---

## Traditional Tools vs. TFDrift-Falco

Traditional drift detection tools (like `driftctl` or `tfsec`) perform **periodic static scans** — they compare Terraform state with cloud reality on a schedule.

TFDrift-Falco takes a fundamentally different approach:

| Aspect | Traditional (Periodic Scan) | TFDrift-Falco (Event-Driven) |
|--------|---------------------------|------------------------------|
| **Detection** | Minutes to hours | Seconds |
| **Who changed it?** | Unknown | Full user identity (IAM, CloudTrail) |
| **When?** | Approximate (last scan window) | Exact timestamp |
| **How?** | Unknown | CloudTrail/Audit Log event details |
| **Mechanism** | `terraform plan` / API polling | Falco gRPC real-time stream |

---

## In One Sentence

> Placing Falco between your infrastructure means **adding a witness to your cloud**.

That's what TFDrift-Falco does — it connects the **blueprint** (Terraform) with the **witness** (Falco) to give you real-time, attributed drift detection.

---

*[Back to README](../README.md)*
