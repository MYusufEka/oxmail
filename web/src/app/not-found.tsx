import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4">
      <h2 className="text-2xl font-bold text-foreground">404</h2>
      <p className="text-sm text-muted-foreground">
        This page could not be found.
      </p>
      <Link
        href="/"
        className="text-sm font-medium text-primary hover:text-primary/80 transition-colors"
      >
        Go back to Dashboard
      </Link>
    </div>
  );
}
