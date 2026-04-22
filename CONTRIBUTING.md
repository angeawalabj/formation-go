# Contribuer à la formation Go

Merci de l'intérêt que tu portes à cette formation. Toutes les contributions sont les bienvenues — qu'il s'agisse d'une faute de frappe, d'une erreur dans un exemple de code, ou d'une suggestion pédagogique.

---

## Ce qu'on accepte volontiers

- **Corrections de fautes** (orthographe, grammaire, ponctuation)
- **Corrections d'erreurs dans le code** (bug, comportement incorrect, import manquant)
- **Améliorations des commentaires** dans le code Go des projets
- **Clarifications pédagogiques** (une explication peu claire, une analogie qui ne fonctionne pas)
- **Problèmes de compatibilité** (la commande X ne fonctionne pas sur Windows / macOS)

## Ce qu'on n'accepte pas pour l'instant

- Ajout de nouveaux chapitres ou modules entiers — la structure est fixe pour la v1
- Traduction dans d'autres langues — pas dans le scope actuel
- Modification de la philosophie pédagogique (l'approche "problème d'abord" est intentionnelle)

---

## Comment proposer une correction

### Pour une correction simple (faute, petit bug)

1. Ouvre une [Issue](https://github.com/angeawalabj/formation-go/issues) en décrivant le problème et son emplacement (fichier + ligne si possible)
2. Ou directement une Pull Request si tu es à l'aise avec Git

### Pour une correction plus importante

1. Ouvre d'abord une Issue pour en discuter
2. Attends un retour avant de travailler dessus
3. Puis soumets une Pull Request

---

## Processus pour une Pull Request

```bash
# 1. Fork le dépôt sur GitHub

# 2. Clone ton fork
git clone https://github.com/TON-PSEUDO/formation-go
cd formation-go

# 3. Crée une branche descriptive
git checkout -b fix/module-02-import-manquant

# 4. Fais tes modifications

# 5. Si tu modifies du code Go, vérifie que les tests passent
cd projets/gowatch && go test -race ./...
cd ../gohub && go test -race ./...

# 6. Commit avec un message clair
git commit -m "fix(module-02): ajouter import manquant dans exemple chapitre 2.3"

# 7. Push et ouvre une Pull Request sur GitHub
git push origin fix/module-02-import-manquant
```

---

## Convention de nommage des commits

```
fix(module-N): description courte       ← Correction d'erreur
docs(module-N): description courte      ← Amélioration de texte
code(projets): description courte       ← Modification du code Go
ci: description courte                  ← CI/GitHub Actions
```

---

## Questions ?

Ouvre une [Issue](https://github.com/angeawalabj/formation-go/issues) avec le label `question`.

Merci de contribuer à rendre cette formation meilleure pour tous.