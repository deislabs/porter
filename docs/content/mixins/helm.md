---
title: helm mixin
description: Manage a Helm release with the helm CLI
---

<img src="/images/mixins/helm.svg" class="mixin-logo" style="width: 150px"/>

Manage a Helm release with the [helm CLI](https://helm.sh).

Source: https://github.com/getporter/helm-mixin

### Install or Upgrade
```
porter mixin install helm
```

### Examples

Install

```yaml
install:
- helm:
    description: "Install MySQL"
    name: mydb
    chart: stable/mysql
    version: 0.10.2
    namespace: mydb
    replace: true
    set:
      mysqlDatabase: wordpress
      mysqlUser: wordpress
    outputs:
    - name: mysql-root-password
      secret: "{{ bundle.parameters.mysql-name }}"
      key: mysql-root-password
    - name: mysql-password
      secret: "{{ bundle.parameters.mysql-name }}"
      key: mysql-password
```

Uninstall

```yaml
uninstall:
- helm:
    description: "Uninstall MySQL"
    purge: true
    releases:
      - mydb
```