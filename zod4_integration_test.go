package supervillain

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZod4ActualCompatibility tests that generated schemas actually work with Zod 4 beta.
// This test creates a temporary TypeScript project, installs Zod 4, and validates the generated code.
func TestZod4ActualCompatibility(t *testing.T) {
	// Skip this test in CI environments or if Node.js is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping TypeScript compilation test")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "supervillain-zod4-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create package.json
	packageJSON := `{
  "name": "supervillain-zod4-test",
  "version": "1.0.0",
  "scripts": {
    "test": "node test.js"
  }
}`
	err = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJSON), 0644)
	require.NoError(t, err)

	// Create TypeScript config
	tsConfig := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  }
}`
	err = os.WriteFile(filepath.Join(tempDir, "tsconfig.json"), []byte(tsConfig), 0644)
	require.NoError(t, err)

	// Install Zod 4 and TypeScript
	cmd := exec.Command("npm", "install", "zod@beta", "typescript", "@types/node")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("npm install output: %s", output)
		t.Skip("Failed to install dependencies, skipping test")
	}

	// Generate complex schema using Supervillain
	type Address struct {
		Street   string  `json:"street"`
		City     string  `json:"city"`
		Country  string  `json:"country"`
		ZipCode  *string `json:"zip_code,omitempty"`
		Verified bool    `json:"verified"`
	}

	type User struct {
		ID          int                `json:"id"`
		Name        string             `json:"name"`
		Email       *string            `json:"email,omitempty"`
		Age         *int               `json:"age,omitempty"`
		Address     *Address           `json:"address,omitempty"`
		Tags        []string           `json:"tags"`
		Preferences map[string]string  `json:"preferences"`
		Scores      map[string]float64 `json:"scores"`
		Active      bool               `json:"active"`
		Balance     float64            `json:"balance"`
	}

	schema := StructToZodSchema(User{})

	// Create TypeScript test file
	testFile := `import { z } from "zod";

` + schema + `

// Test validation with valid data
const validUser: User = {
  id: 1,
  name: "John Doe",
  email: "john@example.com",
  age: 30,
  address: {
    street: "123 Main St",
    city: "New York",
    country: "USA",
    zip_code: "10001",
    verified: true
  },
  tags: ["developer", "typescript"],
  preferences: { "theme": "dark", "lang": "en" },
  scores: { "coding": 95.5, "design": 80.0 },
  active: true,
  balance: 1234.56
};

// Validate the data
try {
  const result = UserSchema.parse(validUser);
  console.log("✅ Validation successful");
  console.log("User ID:", result.id);
  console.log("User Name:", result.name);
  console.log("Address City:", result.address?.city);
  console.log("Tags:", result.tags);
  console.log("Preferences:", result.preferences);
  console.log("Scores:", result.scores);
} catch (error) {
  console.error("❌ Validation failed:", error);
  process.exit(1);
}

// Test with invalid data (should fail)
try {
  const invalidUser = { ...validUser, age: "not a number" };
  UserSchema.parse(invalidUser);
  console.error("❌ Should have failed validation");
  process.exit(1);
} catch (error) {
  console.log("✅ Correctly rejected invalid data");
}

// Test optional fields
const minimalUser: User = {
  id: 2,
  name: "Jane Doe",
  tags: [],
  preferences: {},
  scores: {},
  active: false,
  balance: 0
};

try {
  const result = UserSchema.parse(minimalUser);
  console.log("✅ Minimal user validation successful");
} catch (error) {
  console.error("❌ Minimal user validation failed:", error);
  process.exit(1);
}

console.log("🎉 All Zod 4 compatibility tests passed!");
`

	err = os.WriteFile(filepath.Join(tempDir, "test.ts"), []byte(testFile), 0644)
	require.NoError(t, err)

	// Compile TypeScript
	cmd = exec.Command("npx", "tsc", "test.ts")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("TypeScript compilation failed: %s", output)
		t.Fatalf("Generated schema failed to compile with Zod 4: %v", err)
	}

	// Run the test
	cmd = exec.Command("node", "test.js")
	cmd.Dir = tempDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Generated schema failed to run with Zod 4: %s", output)

	// Verify output contains success messages
	outputStr := string(output)
	assert.Contains(t, outputStr, "✅ Validation successful", "Should validate correct data")
	assert.Contains(t, outputStr, "✅ Correctly rejected invalid data", "Should reject invalid data")
	assert.Contains(t, outputStr, "✅ Minimal user validation successful", "Should validate minimal data")
	assert.Contains(t, outputStr, "🎉 All Zod 4 compatibility tests passed!", "Should complete all tests")

	t.Logf("Zod 4 test output:\n%s", outputStr)
}