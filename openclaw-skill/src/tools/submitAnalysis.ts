import { submitAnalysis as submitToServer } from "../services/serverApi";
import { SubmitAnalysisInput } from "../types";

export async function submitAnalysis(input: SubmitAnalysisInput): Promise<{ status: string }> {
  const tokenAddress = input.tokenAddress.toLowerCase();
  await submitToServer(tokenAddress, input.analysis);
  return { status: "ok" };
}
