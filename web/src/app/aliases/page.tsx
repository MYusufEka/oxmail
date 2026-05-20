import { ArrowLeftRight, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function AliasesPage() {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <ArrowLeftRight className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Aliases</h2>
      </div>
      <div className="flex flex-col items-center justify-center rounded-lg border border-border bg-card p-12">
        <div className="flex size-12 items-center justify-center rounded-full bg-muted">
          <ArrowLeftRight className="size-6 text-muted-foreground" />
        </div>
        <h3 className="mt-4 text-base font-medium text-foreground">
          No aliases yet
        </h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Create email aliases to forward mail between addresses.
        </p>
        <Button className="mt-6" disabled>
          <Plus className="size-4" />
          Add Alias
        </Button>
      </div>
    </div>
  );
}
