	.data
a.0:		.word	0
b.1:		.word	0

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
	li	$3, 4		# a.0 -> $3
	li	$5, 6		# b.1 -> $5
	li	$2, 1
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, b.1
	li	$2, 10
	syscall
	.end main
