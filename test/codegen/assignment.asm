	.data
a.0:		.word	0
b.1:		.word	0
c.2:		.word	0
d.3:		.word	0
e.4:		.word	0
f.5:		.word	0
g.6:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	li	$5, 3		# b.1 -> $5
	sw	$5, b.1		# spilled b.1, freed $5
	li	$5, 2		# c.2 -> $5
	sw	$5, c.2		# spilled c.2, freed $5
	li	$5, 4		# d.3 -> $5
	sw	$5, d.3		# spilled d.3, freed $5
	li	$5, 4		# e.4 -> $5
	sw	$5, e.4		# spilled e.4, freed $5
	li	$5, 8		# f.5 -> $5
	sw	$5, f.5		# spilled f.5, freed $5
	li	$5, 0		# g.6 -> $5
	move	$5, $3		# g.6 -> $5
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, g.6
	li	$2, 10
	syscall
	.end main
