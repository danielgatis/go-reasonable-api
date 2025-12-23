import { render } from "@react-email/components";
import * as fs from "fs";
import * as path from "path";

const TEMPLATES_DIR = path.join(__dirname, "templates");
const OUTPUT_DIR = path.join(__dirname, "../templates");

function toKebabCase(str: string): string {
  return str.replace(/([a-z])([A-Z])/g, "$1-$2").toLowerCase();
}

async function exportTemplates() {
  // Ensure output directory exists
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }

  // Find all template files
  const files = fs.readdirSync(TEMPLATES_DIR).filter((file) => {
    return file.endsWith(".tsx") && file !== "index.ts";
  });

  console.log("Exporting email templates...\n");

  for (const file of files) {
    const templateName = file.replace(".tsx", "");
    const outputName = toKebabCase(templateName);

    // Dynamic import
    const module = await import(`./templates/${templateName}`);
    const Component = module[templateName] || module.default;

    if (!Component) {
      console.error(`  ✗ ${file} - No component found`);
      continue;
    }

    const html = await render(Component());
    const outputPath = path.join(OUTPUT_DIR, `${outputName}.html`);

    fs.writeFileSync(outputPath, html);
    console.log(`  ✓ ${outputName}.html`);
  }

  console.log(`\nDone! Templates exported to ${OUTPUT_DIR}`);
}

exportTemplates().catch(console.error);
