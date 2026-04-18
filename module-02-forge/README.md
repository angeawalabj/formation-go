# Module 02 — La Forge

> *"La perfection est atteinte non pas quand il n'y a plus rien à ajouter, mais quand il n'y a plus rien à retirer."*
> — Antoine de Saint-Exupéry (une phrase que Rob Pike aurait pu écrire)

---

## Ce que ce module va changer

Le Module 01 vous a donné la vision. Le Module 02 vous donne les outils.

Mais attention — ce module ne ressemble pas à un cours de syntaxe classique. On ne va pas lister les types de données Go comme on lirait un dictionnaire. On va les rencontrer là où ils sont utiles, dans des situations concrètes, au moment précis où ils résolvent un vrai problème.

À la fin de ce module, vous ne "connaîtrez" pas la syntaxe Go — vous la **penserez** naturellement.

---

## Les objectifs de ce module

À la fin du Module 02, vous serez capable de :

- ✅ Déclarer et utiliser des variables avec le style idiomatique Go
- ✅ Manipuler des Slices et des Maps sans fuites mémoire
- ✅ Écrire des boucles et des conditions lisibles et efficaces
- ✅ Gérer les erreurs à la façon Go — explicitement, sans Try/Catch
- ✅ Modéliser des données avec des Structs sans la lourdeur de la POO
- ✅ Comprendre et utiliser les interfaces Go pour écrire du code flexible

---

## Ce que vous allez construire

> 🛠️ **Projet fil rouge — Étape 2 : `gowatch` collecte des métriques**

À la fin de ce module, `gowatch` sera capable de :

- Collecter les métriques CPU, RAM et disque de votre machine
- Structurer ces données proprement avec des Structs
- Les afficher en tableau formaté ou en JSON selon un flag CLI
- Gérer les erreurs de façon robuste sans faire planter le programme

```bash
$ ./gowatch --format json
{
  "cpu_count": 8,
  "ram_total_mb": 16384,
  "ram_used_mb": 9821,
  "disk_total_gb": 512,
  "disk_used_gb": 234
}
```

---

## Les chapitres de ce module

### [Chapitre 2.1 — Les briques fondamentales](./01-briques-fondamentales.md)
Variables, types, Slices, Maps. Les fondations du langage expliquées par ce qu'elles permettent de faire — pas par leur définition formelle.

**Concepts abordés :** Variables, inférence de type, Slices, Maps, gestion mémoire.

---

### [Chapitre 2.2 — La logique sans fioritures](./02-logique-sans-fioritures.md)
Go a une seule boucle. Go n'a pas de Try/Catch. Ces deux "manques" sont en réalité deux des décisions les plus intelligentes du langage.

**Concepts abordés :** `for`, conditions, gestion d'erreurs explicite, `defer`, `panic`, `recover`.

---

### [Chapitre 2.3 — Composition vs Héritage](./03-composition-vs-heritage.md)
Comment modéliser le monde réel sans classes, sans héritage, et sans perdre en flexibilité. La réponse de Go est surprenante — et elle fonctionne mieux qu'on ne le croit.

**Concepts abordés :** Structs, méthodes, interfaces, embedding, Duck Typing statique.

---

## Durée estimée

| Chapitre | Lecture | Pratique | Total |
|----------|---------|----------|-------|
| 2.1 — Les briques fondamentales | 25 min | 20 min | 45 min |
| 2.2 — La logique sans fioritures | 25 min | 20 min | 45 min |
| 2.3 — Composition vs Héritage | 30 min | 25 min | 55 min |
| **Total Module 02** | **80 min** | **65 min** | **~2h30** |

---

## L'état d'esprit à adopter

Ce module est celui où beaucoup de développeurs expérimentés résistent.

"Pourquoi Go n'a pas de classes ?"
"Pourquoi je dois écrire `if err != nil` partout ?"
"Pourquoi il n'y a qu'une seule boucle ?"

Ces questions sont légitimes. Et chacune a une réponse solide. L'objectif de ce module n'est pas de vous convaincre que Go a raison sur tout — c'est de vous faire comprendre **pourquoi** ces choix ont été faits, pour que vous puissiez les évaluer honnêtement.

Réservez votre jugement jusqu'à la fin du module. Vous serez surpris.

---

<div align="center">

[⬅️ Retour au Module 01](../module-01-eveil/README.md) · [👉 Commencer le Chapitre 2.1](./01-briques-fondamentales.md)

</div>
