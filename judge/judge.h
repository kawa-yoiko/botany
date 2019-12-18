#ifndef __BOTANY_JUDGE_H__
#define __BOTANY_JUDGE_H__

#include <stdbool.h>

extern const char *judge_chroot;

void compile(const char *sid, const char *lang, const char *contents);
bool is_compiled(const char *sid);
void match(const char *mid, int num_parties, const char *parties[]);

#endif