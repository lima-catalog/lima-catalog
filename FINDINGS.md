# Lima Template Search Findings

## Summary

Our experiments revealed important insights about Lima templates on GitHub and how GitHub's search API works.

## Key Findings

### 1. GitHub Code Search and Forks

**Critical Discovery**: GitHub Code Search **does NOT include forks by default**

- Forks are only searchable if they have more stars than the parent repository
- `lima-vm/lima` has **18,903 stars**
- No fork has more stars, so **all 750 forks are invisible to code search**

### 2. Search Results Breakdown

Our search query: `minimumLimaVersion extension:yml OR extension:yaml -repo:lima-vm/lima`

Results:
- **57 unique templates** found
- **35 unique repositories**
- **34 unique owners**
- **0 from lima-vm/lima** (explicitly excluded)
- **0 from forks** (invisible to search due to star count)

These 57 templates are entirely from **independent repositories** - not lima-vm/lima and not its forks.

### 3. Lima Template Structure

**lima-vm/lima official templates**:
- Located in `templates/` directory (not `examples/`)
- 52 YAML template files
- All contain `minimumLimaVersion` field
- Examples: `ubuntu.yaml`, `k3s.yaml`, `default.yaml`

**Template format**:
```yaml
minimumLimaVersion: 2.0.0

base: template:_images/ubuntu-lts

images: []
cpus: null
memory: null
# ... provisioning scripts, mounts, etc
```

### 4. Independent Templates (57 found)

Examples of repositories with their own Lima templates:
- `felix-kaestner/lima-templates` - Collection of custom templates
- `annie444/utils-util` - Utility templates
- `Yolean/ystack` - Development stack templates
- `danymat/pentest-lab` - Pentesting lab templates
- `0xzer0x/k8s-storage-providers` - K8s storage templates
- `argonprotocol/apps` - App development templates

These represent real community templates worth cataloging!

## Implications for Our Catalog

### What We Should Include

1. ✅ **Independent templates (57 found)** - These are the most valuable
   - Real community use cases
   - Diverse purposes (K8s, dev envs, testing, etc.)
   - Not duplicates of official templates

2. ⚠️ **lima-vm/lima forks (750 forks)** - Need to decide
   - Most are probably unmodified clones
   - Some may have custom templates
   - Would require checking all 750 forks individually

3. ✅ **Official templates (52 in lima-vm/lima)** - For completeness
   - Useful as reference baseline
   - Could mark as "official" in catalog
   - Users might want to compare official vs community

### Recommended Approach

Given the findings, I suggest a **three-tier strategy**:

**Tier 1: Community Templates (Priority)**
- Search for all templates outside lima-vm/lima: ✅ **57 found**
- These are the main value of the catalog
- Fast to collect (< 100 API calls)

**Tier 2: Official Templates (Baseline)**
- Enumerate lima-vm/lima templates: **52 templates**
- Mark as "official" in catalog
- Provides context and comparison

**Tier 3: Modified Forks (Optional/Future)**
- Check all 750 forks of lima-vm/lima
- Compare template SHAs with parent
- Only include if templates differ
- This is expensive (750+ API calls) and low value (most unchanged)

## Scale Assessment

The catalog is **very manageable**:

### Minimum Viable Catalog (Tiers 1 + 2)
- ~110 total templates (57 community + 52 official)
- ~35 unique repos (community only)
- ~34 unique owners (community only)
- API calls needed: ~70-100
- Time to collect: ~2-5 minutes
- Storage: < 50 KB total

### With Fork Checking (Tier 3)
- Potentially 100-200 more templates (unknown until checked)
- 750 additional API calls (one per fork)
- Time: ~15-30 minutes (rate limit dependent)
- Value: Low (most forks are unmodified)

## Recommendations

1. **Start with Tiers 1 + 2**: Focus on community + official templates
   - Delivers immediate value
   - Fast to implement
   - Covers most use cases

2. **Defer Tier 3**: Skip fork checking initially
   - Can add later if there's demand
   - Most forks are unmodified clones
   - Expensive to check relative to value

3. **Search Strategy**: Use simple code search
   - `minimumLimaVersion extension:yml` - finds community templates
   - Direct API access for official templates
   - Skip fork enumeration for now

4. **Filtering**: Post-process search results
   - Filter out lima-vm/lima (if included in results)
   - Check `is_fork` field in metadata
   - For forks, could optionally check if they're forks of lima-vm/lima

## Updated Plan

The original plan is sound, but we should:

1. ✅ Keep the three-tier data structure (templates/repos/orgs)
2. ✅ Keep JSON Lines format
3. ✅ Keep rate limit management
4. ✅ Keep resumability features
5. ✅ Focus on community templates (Tier 1) + official (Tier 2)
6. ⏸️ Defer fork checking (Tier 3) as optional future enhancement

The catalog will be smaller than initially thought, which is **excellent news**:
- Faster to build
- Easier to maintain
- Simpler to browse
- Still captures the valuable community contributions

## Next Steps

1. Implement tool with Tiers 1 + 2
2. Collect and analyze the 57 community templates
3. See what categories/use cases emerge
4. Consider fork checking only if analysis shows value
