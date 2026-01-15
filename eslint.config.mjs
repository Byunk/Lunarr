import eslint from "@eslint/js";
import prettier from "eslint-config-prettier/flat";
import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import tseslint from "typescript-eslint";

export default defineConfig([
  globalIgnores([
    "**/.cache/**",
    "**/.env.*",
    "**/.env",
    "**/*.config.{js,mjs,cjs,ts}",
    "**/build/**",
    "**/coverage/**",
    "**/dist/**",
    "**/node_modules/**",
    "**/public/**",
  ]),
  eslint.configs.recommended,
  tseslint.configs.recommendedTypeChecked,
  prettier,
  {
    settings: {
      react: {
        version: "detect",
      },
    },
    languageOptions: {
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },
        // @ts-ignore
        tsconfigRootDir: import.meta.dirname,
        projectService: {
          allowDefaultProject: ["*.js", "*.mjs", "*.cjs"],
          noWarnOnMultipleProjects: true,
        },
      },
    },
  },
  {
    files: ["**/*.{ts,tsx}"],
    rules: {
      "@typescript-eslint/no-unsafe-assignment": "warn",
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          args: "all",
          argsIgnorePattern: "^_",
          caughtErrors: "all",
          caughtErrorsIgnorePattern: "^_",
          destructuredArrayIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          ignoreRestSiblings: true,
        },
      ],
      "no-empty-pattern": "warn",
    },
  },
  {
    files: ["**/*.{js,mjs,cjs,jsx}"],
    extends: [tseslint.configs.disableTypeChecked],
  },
  {
    files: ["**/*.cjs"],
    languageOptions: {
      globals: {
        ...globals.node,
      },
    },
    rules: {
      "@typescript-eslint/no-require-imports": "off",
    },
  },
]);
