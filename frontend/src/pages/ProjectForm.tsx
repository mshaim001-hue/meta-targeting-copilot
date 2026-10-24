import { useState, type FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Input, Label } from "@/components/ui/input"
import { api } from "@/lib/api"
import { emptyAssets, type Assets } from "@/lib/types"

const assetFields: Array<{ key: keyof Assets; label: string }> = [
  { key: "customerList", label: "Customer list" },
  { key: "pixel", label: "Meta Pixel" },
  { key: "capi", label: "Conversions API" },
  { key: "websiteVisitors", label: "Website visitors" },
  { key: "instagram", label: "Instagram account" },
  { key: "facebookPage", label: "Facebook Page" },
  { key: "videoViews", label: "Video views" },
  { key: "leadForms", label: "Lead forms" },
  { key: "purchasers", label: "Existing purchasers" },
  { key: "crm", label: "CRM" },
]

export function ProjectFormPage() {
  const nav = useNavigate()
  const [error, setError] = useState("")
  const [assets, setAssets] = useState<Assets>(emptyAssets())
  const [pending, setPending] = useState(false)

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const fd = new FormData(e.currentTarget)
    setPending(true)
    setError("")
    try {
      const p = await api.createProject({
        businessName: fd.get("businessName"),
        website: fd.get("website"),
        product: fd.get("product"),
        country: fd.get("country"),
        city: fd.get("city"),
        price: fd.get("price"),
        objective: fd.get("objective"),
        budget: fd.get("budget"),
        assets,
      })
      nav(`/projects/${p.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed")
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="max-w-3xl">
      <p className="text-xs uppercase tracking-[0.2em] text-ink/50">Screen 1</p>
      <h1 className="mt-2 font-serif text-4xl">Project</h1>
      <form className="mt-8 space-y-6" onSubmit={onSubmit}>
        <Card className="grid gap-4 md:grid-cols-2">
          <Field name="businessName" label="Business name" required />
          <Field name="website" label="Website" />
          <Field name="product" label="Product / service" required />
          <Field name="objective" label="Conversion goal" placeholder="Memberships / leads" />
          <Field name="country" label="Country" required defaultValue="Kazakhstan" />
          <Field name="city" label="City" required defaultValue="Astana" />
          <Field name="price" label="Price" />
          <Field name="budget" label="Monthly ad budget" />
        </Card>
        <Card>
          <p className="text-sm font-medium">First-party data</p>
          <p className="mt-1 text-sm text-ink/60">Used only for retargeting / lookalike planning. Nothing is pushed to Meta Ads from this app.</p>
          <div className="mt-4 grid gap-2 md:grid-cols-2">
            {assetFields.map((f) => (
              <label key={f.key} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={assets[f.key]}
                  onChange={(e) => setAssets({ ...assets, [f.key]: e.target.checked })}
                />
                {f.label}
              </label>
            ))}
          </div>
        </Card>
        {error ? <p className="text-sm text-missing">{error}</p> : null}
        <Button type="submit" disabled={pending}>Continue to audience prompt</Button>
      </form>
    </div>
  )
}

function Field(props: { name: string; label: string; required?: boolean; defaultValue?: string; placeholder?: string }) {
  return (
    <div className="space-y-1">
      <Label htmlFor={props.name}>{props.label}</Label>
      <Input id={props.name} name={props.name} required={props.required} defaultValue={props.defaultValue} placeholder={props.placeholder} />
    </div>
  )
}
