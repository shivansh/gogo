# Test to find n'th fibonacci number

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
retVal:		.word	0
str:		.asciiz "n'th fibonacci number: "
first:		.word	0
second:		.word	0
temp:		.word	0

	.text


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	sw	$3, n		# spilled n, freed $3
	li	$3, 0		# i -> $3
	sw	$3, i
	jal	fib
	lw	$3, i
	sw	$3, i		# spilled i, freed $3
	move	$3, $2
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	move	$4, $3
	syscall
	# Store variables back into memory
	sw	$3, retVal
	li	$2, 10
	syscall
	.end main

	.globl fib
	.ent fib
fib:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$3, 1		# first -> $3
	sw	$3, first	# spilled first, freed $3
	li	$3, 1		# second -> $3
	sw	$3, second	# spilled second, freed $3
	lw	$3, n
	sub	$3, $3, 2	# n -> $3
	# Store variables back into memory
	sw	$3, n

loop:
	lw	$3, n
	ble	$3, 0, exit
	lw	$5, second
	move	$6, $5		# temp -> $6
	lw	$7, first
	add	$5, $5, $7	# second -> $5
	move	$7, $6		# first -> $7
	sub	$3, $3, 1	# n -> $3
	# Store variables back into memory
	sw	$3, n
	sw	$5, second
	sw	$6, temp
	sw	$7, first
	j	loop

exit:
	lw	$2, second

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end fib
