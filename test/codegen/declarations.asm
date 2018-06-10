	.data
return.0:	.word	0
return.1:	.word	0
b.0:		.word	0
d.1:		.word	0
x.2:		.word	0
a.3:		.word	0
c.4:		.word	0
e.5:		.word	0
f.6:		.word	0
g.7:		.word	0

	.text
temp:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$3, 1		# return.0 -> $3
	sw	$3, return.0	# spilled return.0, freed $3
	li	$3, 2		# return.1 -> $3
	# Store dirty variables back into memory
	sw	$3, return.1

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end temp

	.globl main
	.ent main
main:
	li	$3, 6		# b.0 -> $3
	sw	$3, b.0		# spilled b.0, freed $3
	li	$3, 3		# d.1 -> $3
	li	$5, 3		# x.2 -> $5
	sw	$5, x.2		# spilled x.2, freed $5
	move	$5, $3		# a.3 -> $5
	li	$2, 1
	move	$4, $5
	syscall
	sw	$3, d.1
	sw	$5, a.3
	jal	temp
	lw	$3, d.1
	lw	$5, a.3
	lw	$6, return.0	# return.0 -> $6
	move	$7, $6		# c.4 -> $7
	sw	$7, c.4		# spilled c.4, freed $7
	lw	$7, return.1	# return.1 -> $7
	move	$8, $7		# e.5 -> $8
	sw	$8, e.5		# spilled e.5, freed $8
	li	$8, 0		# f.6 -> $8
	sw	$8, f.6		# spilled f.6, freed $8
	li	$8, 0		# g.7 -> $8
	sw	$3, d.1
	sw	$5, a.3
	sw	$6, return.0
	sw	$7, return.1
	sw	$8, g.7
	jal	temp
	lw	$3, d.1
	lw	$5, a.3
	lw	$8, g.7
	move	$8, $6		# f.6 -> $8
	move	$9, $7		# g.7 -> $9
	li	$2, 1
	move	$4, $8
	syscall
	li	$2, 1
	move	$4, $9
	syscall
	# Store dirty variables back into memory
	sw	$8, f.6
	sw	$9, g.7
	li	$2, 10
	syscall
	.end main
