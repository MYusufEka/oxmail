"use client";

import { useState, useMemo, useCallback } from "react";
import { useAuth } from "@/contexts/auth";
import { redirect } from "next/navigation";
import { Filter, Plus, Trash2, Save, Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useSieveScript, useSetSieveScript } from "@/hooks/use-sieve";
import { useMailFolders } from "@/hooks/use-mail";
import {
  generateFilterScript,
  generateVacationScript,
  parseFilterRules,
  parseVacationSettings,
  mergeSieveScripts,
  type FilterRule,
  type FilterField,
  type FilterOperator,
  type FilterAction,
} from "@/lib/sieve-utils";

interface AddFilterDialogProps {
  folders: string[];
  onSave: (rule: FilterRule) => void;
}

function AddFilterDialog({ folders, onSave }: AddFilterDialogProps) {
  const [open, setOpen] = useState(false);
  const [field, setField] = useState<FilterField>("from");
  const [operator, setOperator] = useState<FilterOperator>("contains");
  const [value, setValue] = useState("");
  const [action, setAction] = useState<FilterAction>("move");
  const [targetFolder, setTargetFolder] = useState("INBOX");

  function handleSave() {
    if (!value.trim()) {
      toast.error("Enter a filter value");
      return;
    }
    onSave({
      id: `rule-${Date.now()}`,
      field,
      operator,
      value: value.trim(),
      action,
      targetFolder: action === "move" ? targetFolder : undefined,
    });
    setValue("");
    setOpen(false);
  }

  function handleOpenChange(next: boolean) {
    if (!next) {
      setField("from");
      setOperator("contains");
      setValue("");
      setAction("move");
      setTargetFolder("INBOX");
    }
    setOpen(next);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button data-testid="add-filter-trigger">
          <Plus className="size-4" />
          Add Filter
        </Button>
      </DialogTrigger>
      <DialogContent data-testid="add-filter-dialog">
        <DialogHeader>
          <DialogTitle>Add Filter Rule</DialogTitle>
          <DialogDescription>
            Route or discard incoming mail based on message headers.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Field</label>
              <Select value={field} onValueChange={(v: FilterField) => setField(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="from">From</SelectItem>
                  <SelectItem value="to">To</SelectItem>
                  <SelectItem value="subject">Subject</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Operator</label>
              <Select value={operator} onValueChange={(v: FilterOperator) => setOperator(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="contains">Contains</SelectItem>
                  <SelectItem value="is">Is</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted-foreground">Value</label>
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder="e.g. newsletter@example.com"
              autoFocus
              data-testid="filter-value-input"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">Action</label>
              <Select value={action} onValueChange={(v: FilterAction) => setAction(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="move">Move to folder</SelectItem>
                  <SelectItem value="delete">Delete</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {action === "move" && (
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted-foreground">Target folder</label>
                <Select value={targetFolder} onValueChange={setTargetFolder}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {folders.map((f) => (
                      <SelectItem key={f} value={f}>{f}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={handleSave} data-testid="add-filter-save">Add Rule</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function FiltersPage() {
  const { user, isLoading: authLoading } = useAuth();
  const userEmail = user?.email ?? "";

  const { data: sieveData, isLoading: sieveLoading, isError, refetch } = useSieveScript(userEmail);
  const { data: foldersData } = useMailFolders(userEmail);
  const setSieveMutation = useSetSieveScript();

  const parsedRules = useMemo(
    () => parseFilterRules(sieveData?.script ?? ""),
    [sieveData?.script],
  );
  const [rulesDraft, setRulesDraft] = useState<FilterRule[] | null>(null);
  const rules = rulesDraft ?? parsedRules;

  const folders = useMemo(() => {
    const apiFolders = foldersData?.folders ?? [];
    const knownSystem = new Set(["INBOX", "Sent", "Drafts", "Trash", "Junk", "Archive"]);
    const all = apiFolders.map((f) => f.name);
    const system = all.filter((f) => knownSystem.has(f));
    const custom = all.filter((f) => !knownSystem.has(f));
    return [...system, ...custom];
  }, [foldersData]);

  const handleAddRule = useCallback((rule: FilterRule) => {
    setRulesDraft((previousRules) => [...(previousRules ?? parsedRules), rule]);
  }, [parsedRules]);

  const handleDeleteRule = useCallback((ruleId: string) => {
    setRulesDraft((previousRules) =>
      (previousRules ?? parsedRules).filter((rule) => rule.id !== ruleId),
    );
  }, [parsedRules]);

  const handleSave = useCallback(async () => {
    try {
      const fullScript = sieveData?.script ?? "";
      const vacation = parseVacationSettings(fullScript);
      const vacationScript = generateVacationScript(vacation);
      const filterScript = generateFilterScript(rules);
      const merged = mergeSieveScripts(filterScript, vacationScript);

      await setSieveMutation.mutateAsync({ email: userEmail, script: merged });
      toast.success("Filter rules saved");
    } catch (err) {
      toast.error("Failed to save", { description: err instanceof Error ? err.message : "Unknown error" });
    }
  }, [rules, sieveData, setSieveMutation, userEmail]);

  if (authLoading) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-3">
          <Filter className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Filters</h2>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col gap-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!user) {
    redirect("/login");
  }

  if (sieveLoading) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-3">
          <Filter className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Filters</h2>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col gap-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-3">
          <Filter className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Filters</h2>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center p-12">
            <div className="flex size-12 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="size-6 text-destructive" />
            </div>
            <h3 className="mt-4 text-base font-medium text-foreground">Failed to load filters</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Something went wrong. Please try again.
            </p>
            <Button variant="outline" className="mt-6" onClick={() => refetch()}>
              <RefreshCw className="size-4" />
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Filter className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Filters</h2>
        </div>
        <div className="flex items-center gap-2">
          <AddFilterDialog folders={folders} onSave={handleAddRule} />
          <Button
            onClick={handleSave}
            disabled={setSieveMutation.isPending}
            data-testid="save-filters"
          >
            {setSieveMutation.isPending && <Loader2 className="size-4 animate-spin" />}
            <Save className="size-4" />
            Save
          </Button>
        </div>
      </div>

      {rules.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center p-12">
            <div className="flex size-12 items-center justify-center rounded-full bg-muted">
              <Filter className="size-6 text-muted-foreground" />
            </div>
            <h3 className="mt-4 text-base font-medium text-foreground">No filter rules</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Add rules to automatically sort or discard incoming mail.
            </p>
          </CardContent>
        </Card>
      )}

      {rules.length > 0 && (
        <div className="flex flex-col gap-2">
          {rules.map((rule) => (
            <Card key={rule.id} className="py-3" data-testid={`filter-rule-${rule.id}`}>
              <CardContent className="flex items-center justify-between px-4">
                <div className="flex items-center gap-3 text-sm">
                  <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    {rule.action === "delete" ? "DELETE" : "MOVE"}
                  </span>
                  <span className="text-muted-foreground">If</span>
                  <span className="font-medium">{rule.field}</span>
                  <span className="text-muted-foreground">{rule.operator}</span>
                  <span className="font-medium">{rule.value}</span>
                  {rule.action === "move" && (
                    <>
                      <span className="text-muted-foreground">→</span>
                      <span className="font-medium">{rule.targetFolder}</span>
                    </>
                  )}
                </div>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => handleDeleteRule(rule.id)}
                  aria-label={`Delete rule for ${rule.field} ${rule.operator} ${rule.value}`}
                >
                  <Trash2 className="size-4 text-muted-foreground hover:text-destructive" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
