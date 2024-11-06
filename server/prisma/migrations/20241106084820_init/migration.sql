/*
  Warnings:

  - Made the column `codePath` on table `Repo` required. This step will fail if there are existing NULL values in that column.

*/
-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_Repo" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "name" TEXT NOT NULL,
    "address" TEXT NOT NULL,
    "username" TEXT NOT NULL,
    "pwd" TEXT NOT NULL,
    "codePath" TEXT NOT NULL
);
INSERT INTO "new_Repo" ("address", "codePath", "createdAt", "id", "name", "pwd", "updatedAt", "username") SELECT "address", "codePath", "createdAt", "id", "name", "pwd", "updatedAt", "username" FROM "Repo";
DROP TABLE "Repo";
ALTER TABLE "new_Repo" RENAME TO "Repo";
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
