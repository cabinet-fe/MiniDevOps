/*
  Warnings:

  - You are about to drop the column `branches` on the `Repo` table. All the data in the column will be lost.
  - You are about to drop the column `url` on the `Repo` table. All the data in the column will be lost.
  - Added the required column `address` to the `Repo` table without a default value. This is not possible if the table is not empty.

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
    "codePath" TEXT
);
INSERT INTO "new_Repo" ("codePath", "createdAt", "id", "name", "pwd", "updatedAt", "username") SELECT "codePath", "createdAt", "id", "name", "pwd", "updatedAt", "username" FROM "Repo";
DROP TABLE "Repo";
ALTER TABLE "new_Repo" RENAME TO "Repo";
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
