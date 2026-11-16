# Evidence

## Upstream and integrity

- The [v1.0.0 Intel Dockerfile](https://github.com/ggml-org/llama.cpp/blob/v1.0.0/.devops/intel.Dockerfile) installs oneDNN `2026.1.0.625` alongside Deep Learning Essentials `2026.2.0`.
- [Intel's toolkit compatibility table](https://www.intel.com/content/www/us/en/developer/articles/release-notes/oneapi-toolkit/2026.html) lists standalone oneDNN `2026.1` with toolkit `2026.2`. The [oneDNN release notes](https://www.intel.com/content/www/us/en/developer/articles/release-notes/onednn/2026.html) identify GPU and SYCL Graph changes. No performance improvement is assumed for this deployment.
- The official installer download was 365757060 bytes and had SHA256 `38c2585bc01b62e011cc1678e090be68ed1574c7a9fc9d775b09cb6ebd041427`. This is a hash of the downloaded file, not an independently authenticated vendor checksum.
- The candidate built successfully from llama.cpp `b1b770d90d94cad9e9ea5b993bdbc383a522fb28` and the resolved driver manifest. The build emitted SYCL deprecation warnings but completed.
- The candidate final-image loader resolved oneDNN from `/opt/intel/oneapi/dnnl/2026.1/lib/libdnnl.so.3`, along with Level Zero, SYCL and MKL. Both oneDNN notices were present. Device discovery reported `SYCL0`.
- Both recorded OpenCL ICD targets and `/lib/x86_64-linux-gnu/libze_intel_gpu.so.1` resolved through the final image's ELF loader without missing dependencies. The candidate's local image size was 1,432,399,726 bytes versus 1,431,071,485 bytes for the baseline.
- The oneDNN verbose log identified the candidate library as `v3.13.2` at source commit `c852e73ef73b861f95d3b96fbe7518b330bc24d5`, versus baseline `v3.11.4` at `0291f89430882d350585d276cc0d0eda623906e9`. An upstream source compare reports that neither of these commits contains the [v3.14.2 query-length-16 SDPA fix](https://github.com/oneapi-src/oneDNN/commit/4db284662a37cf46401655b5ee99391d9de52c00). llama.cpp v1.0.0's oneDNN dispatch uses a 32-query-token minimum; this source observation does not prove every downstream path is unaffected.

## Existing deployed image baseline

- The currently deployed v1.0.0 image responds to a bounded Gemma prompt with `42` and a coding prompt with `sum(xs)`.
- The deployed embedding model returned two finite 768-dimensional vectors. These checks do not measure embedding quality or establish full-context fidelity.
- The separately deployed Whisper image passed the pinned JFK transcription check. This change does not modify Whisper.
- After that check, the deployed llama and Whisper runtimes independently exited with code 0 after their idle timeout.

## Candidate comparisons

- In isolated containers, the existing and candidate images returned identical deterministic `42` and `sum(xs)` answers using q8 KV caches.
- Two short embedding vectors were identical across images and each contained 768 finite values.
- A longer embedding fixture was run on both images with `SYCL0`, 25 of 25 model layers offloaded, flash attention and oneDNN enabled. Verbose logs explicitly recorded `FA-DISP ... ONEDNN` and oneDNN GPU SDPA execution. The two 768-element embeddings were identical, with cosine similarity 1.0 and maximum absolute difference 0.
- These are bounded comparisons, not a full-context, multimodal or every-preset fidelity sweep. No performance improvement is inferred.
