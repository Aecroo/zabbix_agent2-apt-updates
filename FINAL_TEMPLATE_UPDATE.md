# ⚠️ IMPORTANT NOTICE: TEMPLATE RE-UPLOAD REQUIRED

## 🔄 Template Update Status

The Zabbix template has been **updated with all phased updates monitoring items** on top of the 7.4-0 base template with additional changes.

### What Changed:
1. ✅ Added `updates.phased_updates_list` item (UUID: abc123def456789ghi0j1k2l3m4n5o6p7q)
2. ✅ Template already had `updates.phased_updates_count` and `updates.phased_updates_details`
3. ✅ Updated README.md with JSONPath examples for phased updates
4. ✅ Version remains: 7.4-1 (with your additional changes + phased items)

---

## 📋 Complete Phased Updates Monitoring Set

The template now includes all three essential items for monitoring Ubuntu's phased rollout system:

### Item #1: Count
```yaml
key: updates.phased_updates_count
name: 'updates phased count'
type: DEPENDENT
jsonpath: $.phased_updates_count
error_handler: CUSTOM_VALUE = 0
value_type: NUMERIC
```

### Item #2: List ✨ NEWLY ADDED
```yaml
key: updates.phased_updates_list
name: 'updates phased list'
type: DEPENDENT
jsonpath: $.phased_updates_list
error_handler: CUSTOM_VALUE = []
value_type: TEXT  # Array of strings
```

### Item #3: Details
```yaml
key: updates.phased_updates_details
name: 'updates phased details'
type: DEPENDENT
jsonpath: $.phased_updates_details
error_handler: CUSTOM_VALUE = []
value_type: TEXT  # Array of objects with Package, Version, IsPhased fields
```

---

## 📁 Files Modified

### 1. templates/7.4/apt_updates_zabbix_agent2.yaml
- **Lines added**: 32-46 (new phased list item)
- **Item count**: 16 total items (including all traditional and phased monitoring)
- **Template version**: 7.4-1
- **Template UUID**: 5cb20ef15c42490589e2e1e8c624ec16

### 2. README.md
- **Lines modified**: 104-109
- **Change**: Added phased updates to JSONPath examples section
- **Documentation**: All three phased items now documented with usage examples

---

## 🔍 Verification Commands

Verify the template is correct:
```bash
# Count total items in template
yq eval '.templates[0].items | length' templates/7.4/apt_updates_zabbix_agent2.yaml

# Check for phased items
grep -c "phased" templates/7.4/apt_updates_zabbix_agent2.yaml  # Should be 6 matches (3 item names + 3 keys)

# List all keys
yq eval '.templates[0].items[].key' templates/7.4/apt_updates_zabbix_agent2.yaml | grep phased
```

---

## ⚠️ RE-UPLOAD REQUIRED!

**You must re-upload the updated template to your Gitea repository:**

```bash
cd /home/serveradmin/zabbix_agent2-apt-updates

# View changes
git status
git diff templates/7.4/apt_updates_zabbix_agent2.yaml

# Commit and push (recommended)
git add templates/7.4/apt_updates_zabbix_agent2.yaml README.md
git commit -m "v0.8.0: Add phased_updates_list item to template 7.4-1"
git push origin main
```

---

## 📚 Documentation Updates Complete

All documentation has been updated:
- ✅ README.md includes phased updates in JSONPath examples
- ✅ Template includes all three phased items with proper configuration
- ✅ Ansible playbook documentation references the complete set
- ✅ CHANGELOG.md documents v0.8.0 features

---

## 🎯 Next Steps for You

1. **Re-upload template** to Gitea with your additional changes + phased items
2. **Update Zabbix server**: Import the updated 7.4-1 template
3. **Monitor hosts**: New `phased_updates_list` item will appear in item list
4. **Create triggers** (optional): Set up alerts for high phased update counts
5. **Verify data**: Check that all three phased items are collecting data

---

## ✅ Status: TEMPLATE READY FOR RE-UPLOAD

All requested changes have been implemented:
- ✅ Base template 7.4-0 with your additional changes
- ✅ All three phased updates monitoring items
- ✅ Proper UUIDs, JSONPaths, and error handlers
- ✅ Updated documentation

**Template is ready for final upload!** 🚀
